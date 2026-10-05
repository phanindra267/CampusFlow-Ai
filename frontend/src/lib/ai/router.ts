/**
 * Deterministic AI request router.
 *
 * This is the first thing every AI request touches, and it is the main reason
 * the application is fast: structured campus operations and ordinary search are
 * answered from PostgreSQL/Weaviate without ever loading a language model.
 *
 * Routing is pure string work — normalisation, keyword and phrase matching, and
 * a small amount of regex. There is no model call, no network I/O and no
 * blocking, so it costs microseconds and stays testable.
 *
 * The router classifies a request into one of three outcomes:
 *
 *   STRUCTURED  a known application operation → call the existing REST endpoint
 *   SEARCH      free-text retrieval           → hybrid BM25 + vector search
 *   GENERATIVE  reasoning or synthesis       → retrieve context, then WebLLM
 */

export type IntentKind = "STRUCTURED" | "SEARCH" | "GENERATIVE";

export type Intent =
  | "EVENT_SEARCH"
  | "CLUB_SEARCH"
  | "OPPORTUNITY_SEARCH"
  | "ANNOUNCEMENT_QUERY"
  | "RESOURCE_BOOKING"
  | "NOTIFICATION_QUERY"
  | "PROFILE_QUERY"
  | "SERVICE_REQUEST_QUERY"
  | "REGISTRATION_QUERY"
  | "APPLICATION_QUERY"
  | "RESEARCH_QUERY"
  | "DASHBOARD_QUERY"
  | "HYBRID_SEARCH"
  | "COMPLEX_QUERY";

export type RouteConfidence = "high" | "medium" | "low";

export type Route = {
  intent: Intent;
  kind: IntentKind;
  confidence: RouteConfidence;
  /** 0..1, derived from match strength. Useful for choosing a presentation. */
  score: number;
  /**
   * The existing REST path that serves this intent. STRUCTURED and SEARCH
   * routes resolve entirely on the backend; GENERATIVE routes use this as the
   * retrieval source for grounding context.
   */
  endpoint: string;
  /** Terms left after intent words are removed, for retrieval. */
  terms: string[];
  /** Free-text query to retrieve with, empty for pure navigation requests. */
  searchQuery: string;
};

/**
 * Words that name an application entity. Presence of one of these is what
 * separates a structured campus request from a reasoning request.
 */
const ENTITY_TERMS: Record<Exclude<Intent, "HYBRID_SEARCH" | "COMPLEX_QUERY">, string[]> = {
  EVENT_SEARCH: ["event", "events", "webinar", "workshop", "hackathon", "seminar", "meetup", "fest"],
  CLUB_SEARCH: ["club", "clubs", "society", "societies", "group", "groups", "chapter", "guild"],
  // "research" is deliberately absent: it names the research domain, not a
  // posting. A research internship still matches on "internship", and
  // "research papers" reaches RESEARCH_QUERY instead.
  OPPORTUNITY_SEARCH: [
    "internship",
    "internships",
    "opportunity",
    "opportunities",
    "job",
    "jobs",
    "role",
    "roles",
    "placement",
    "fellowship",
    "stipend",
    "driving",
  ],
  ANNOUNCEMENT_QUERY: ["announcement", "announcements", "notice", "notices", "circular", "bulletin"],
  RESOURCE_BOOKING: [
    "book",
    "booking",
    "reserve",
    "reservation",
    "room",
    "rooms",
    "lab",
    "labs",
    "library",
    "studio",
    "venue",
    "seat",
    "slot",
    "study",
  ],
  NOTIFICATION_QUERY: ["notification", "notifications", "alerts", "inbox"],
  PROFILE_QUERY: ["profile", "profiles", "my details", "my account", "my info", "about me"],
  SERVICE_REQUEST_QUERY: [
    "ticket",
    "tickets",
    "complaint",
    "complaints",
    "request",
    "requests",
    "support",
    "helpdesk",
    "grievance",
  ],
  REGISTRATION_QUERY: ["registration", "registrations", "registered", "rsvp", "sign up", "signup"],
  APPLICATION_QUERY: ["application", "applications", "applied", "apply"],
  RESEARCH_QUERY: [
    "research",
    "publication",
    "publications",
    "paper",
    "papers",
    "lab",
    "labs",
    "researcher",
    "supervisor",
    "faculty",
  ],
  DASHBOARD_QUERY: ["dashboard", "overview", "summary", "stats", "analytics", "activity"],
};

/**
 * Generative markers, each carrying a weight. A marker is any phrasing that
 * asks for judgement, synthesis, comparison or a plan — things a template
 * cannot produce and a lookup cannot answer.
 *
 * They are checked before entity keywords, so "summarise these opportunities"
 * is treated as reasoning about opportunities rather than a listing. The
 * strongest single marker decides, not their sum: that request mentions several
 * entities but carries one clear signal that generation is what is asked for.
 */
const GENERATIVE_MARKERS: Array<{ pattern: RegExp; weight: number }> = [
  { pattern: /\b(summari[sz]e|recap|tl;?dr)\b/i, weight: 0.9 },
  { pattern: /\b(explain|describe|clarify|teach me|walk me through)\b/i, weight: 0.8 },
  { pattern: /\b(compare|contrast|difference|differences|versus|vs\.?)\b/i, weight: 0.9 },
  { pattern: /\b(recommend|suggest|advice|advise)\b/i, weight: 0.85 },
  { pattern: /\b(which (one|ones|should|do)|what should i|best (way|option|for))\b/i, weight: 0.85 },
  { pattern: /\b(plan|roadmap|strategy|steps?|guide me|help me (start|prepare|learn))\b/i, weight: 0.75 },
  { pattern: /\b(based on|considering|given my|most relevant|relevant to (me|someone))\b/i, weight: 0.8 },
  { pattern: /\b(why|how (do|can|should) i|what should i)\b/i, weight: 0.7 },
  { pattern: /\b(prepare|practi[cs]e|interview|revision)\b/i, weight: 0.65 },
  { pattern: /\b(help me understand|tell me more about|help me decide)\b/i, weight: 0.9 },
];

/**
 * GENERATIVE_THRESHOLD is the strongest-marker weight required to route to the
 * model. It sits above the weaker advice markers so a bare "show me events
 * about internships" stays on the fast path.
 */
const GENERATIVE_THRESHOLD = 0.6;

/**
 * Time scoping. A request that only asks to *list* something stays structured
 * even with a time word attached, because the backend already filters by date.
 */
const TIME_SCOPES = [
  { pattern: /\b(today|tonight|right now|currently)\b/i, scope: "today" },
  { pattern: /\b(this week|coming week|next few days)\b/i, scope: "this_week" },
  { pattern: /\b(this month|coming month)\b/i, scope: "this_month" },
  { pattern: /\b(upcoming|up next|soon|future|next)\b/i, scope: "upcoming" },
  { pattern: /\b(past|last|previous|recent|archive)\b/i, scope: "past" },
];

const POLITE_PREFIX =
  /^(please|kindly|can you|could you|would you|i want to|i need to|i'd like to|show me|list|find|get|display|search for|search)\b\s*/i;

/** Words dropped from a search query because they add no retrieval signal. */
const STOP_WORDS = new Set([
  "a", "an", "the", "for", "of", "in", "on", "at", "to", "and", "or", "me", "my",
  "show", "list", "find", "get", "display", "please", "can", "you", "i", "want",
  "need", "are", "is", "what", "which", "there", "any", "some", "all", "with",
  "about", "that", "this", "have", "has", "do", "does", "tell", "give", "up",
]);

function normalize(raw: string): string {
  return raw
    .toLowerCase()
    .replace(/[’']/g, "'")
    .replace(/[^a-z0-9\s'-]/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

/**
 * scoreEntities counts how strongly the text points at each application
 * entity. Exact phrases score highest, then word matches, so "show events"
 * beats a passing mention of "event" in a longer question.
 */
function scoreEntities(text: string): Map<Intent, number> {
  const scores = new Map<Intent, number>();

  for (const [intent, terms] of Object.entries(ENTITY_TERMS) as Array<
    [Exclude<Intent, "HYBRID_SEARCH" | "COMPLEX_QUERY">, string[]]
  >) {
    let score = 0;

    for (const term of terms) {
      // A multi-word phrase is a stronger signal than a bare keyword, because
      // "my profile" is unambiguously about the profile while "profile" could
      // appear in any sentence.
      if (term.includes(" ")) {
        if (text.includes(` ${term} `)) score += 2;
        continue;
      }
      if (new RegExp(`\\b${escapeRegExp(term)}\\b`).test(text)) score += 1;
    }

    if (score > 0) scores.set(intent, score);
  }

  return scores;
}

/** True when the request is essentially navigation with no free-text content. */
function isBareListing(text: string): boolean {
  return /^(show|list|find|get|display|search|see|browse|open)?\s*(me\s+)?(all\s+|the\s+|my\s+|upcoming\s+|recent\s+)*[a-z]*\s*$/.test(
    text,
  );
}

function extractTimeScope(text: string): string | null {
  for (const { pattern, scope } of TIME_SCOPES) {
    if (pattern.test(text)) return scope;
  }
  return null;
}

/** Pull the searchable terms out of a request once intent words are removed. */
function extractTerms(original: string): string[] {
  const withoutPrefix = original.replace(POLITE_PREFIX, "");
  const words = normalize(withoutPrefix).split(" ").filter(Boolean);

  const meaningful = words.filter((word) => !STOP_WORDS.has(word));
  return meaningful.length > 0 ? meaningful : words;
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function toConfidence(score: number): RouteConfidence {
  if (score >= 2) return "high";
  if (score >= 1) return "medium";
  return "low";
}

/**
 * strongestGenerativeWeight returns the weight of the strongest generative
 * marker present, or 0 when none match. Taking the maximum rather than the sum
 * keeps one clear signal ("summarise…") from being diluted by a second weaker
 * one.
 */
function strongestGenerativeWeight(text: string): number {
  let strongest = 0;
  for (const { pattern, weight } of GENERATIVE_MARKERS) {
    if (weight > strongest && pattern.test(text)) {
      strongest = weight;
    }
  }
  return strongest;
}

/** Deterministic tie-break so identical inputs always route identically. */
const ENDPOINTS: Record<Intent, string> = {
  EVENT_SEARCH: "/events",
  CLUB_SEARCH: "/clubs",
  OPPORTUNITY_SEARCH: "/opportunities",
  ANNOUNCEMENT_QUERY: "/announcements",
  RESOURCE_BOOKING: "/community/resources",
  NOTIFICATION_QUERY: "/community/notifications",
  PROFILE_QUERY: "/auth/me",
  SERVICE_REQUEST_QUERY: "/community/service-requests",
  REGISTRATION_QUERY: "/events",
  APPLICATION_QUERY: "/opportunities",
  RESEARCH_QUERY: "/research",
  DASHBOARD_QUERY: "/dashboard",
  HYBRID_SEARCH: "/community/search",
  COMPLEX_QUERY: "/community/search",
};

/**
 * route classifies a natural-language request.
 *
 * Deterministic and synchronous by design: it must stay cheap enough to run on
 * every keystroke-driven submit without measurable cost.
 */
export function route(request: string): Route {
  const trimmed = request.trim();
  const text = ` ${normalize(trimmed)} `;

  if (trimmed.length === 0) {
    return {
      intent: "COMPLEX_QUERY",
      kind: "GENERATIVE",
      confidence: "low",
      score: 0,
      endpoint: ENDPOINTS.COMPLEX_QUERY,
      terms: [],
      searchQuery: "",
    };
  }

  const entities = scoreEntities(text);
  const strongest = [...entities.entries()].sort((a, b) =>
    b[1] === a[1] ? a[0].localeCompare(b[0]) : b[1] - a[1],
  )[0];

  const terms = extractTerms(trimmed);
  const timeScope = extractTimeScope(text);
  const bare = isBareListing(text);

  // Generation markers win over entity keywords: "summarise these
  // opportunities" is a reasoning request about opportunities, not a listing.
  const generative = strongestGenerativeWeight(trimmed);

  if (generative >= GENERATIVE_THRESHOLD) {
    return {
      intent: "COMPLEX_QUERY",
      kind: "GENERATIVE",
      confidence: generative >= 0.8 ? "high" : "medium",
      score: Number(generative.toFixed(3)),
      endpoint: strongest
        ? ENDPOINTS[strongest[0]]
        : ENDPOINTS.COMPLEX_QUERY,
      terms,
      searchQuery: terms.join(" "),
    };
  }

  if (strongest) {
    const [intent, score] = strongest;

    // A single weak entity mention inside a longer sentence is more likely a
    // free-text search than a structured listing.
    if (score < 1.5 && !bare && terms.length > 4) {
      return {
        intent: "HYBRID_SEARCH",
        kind: "SEARCH",
        confidence: "low",
        score: Number((score / 2).toFixed(3)),
        endpoint: ENDPOINTS.HYBRID_SEARCH,
        terms,
        searchQuery: terms.join(" "),
      };
    }

    return {
      intent,
      kind: "STRUCTURED",
      confidence: toConfidence(score),
      score: Number(Math.min(score / 3, 1).toFixed(3)),
      endpoint: ENDPOINTS[intent],
      terms,
      searchQuery: terms.join(" "),
    };
  }

  // No entity keyword at all, but free text: still a search, not a generation.
  if (terms.length > 0) {
    return {
      intent: "HYBRID_SEARCH",
      kind: "SEARCH",
      confidence: bare ? "medium" : "low",
      score: bare ? 0.5 : 0.3,
      endpoint: ENDPOINTS.HYBRID_SEARCH,
      terms,
      searchQuery: terms.join(" "),
    };
  }

  return {
    intent: "COMPLEX_QUERY",
    kind: "GENERATIVE",
    confidence: "low",
    score: 0,
    endpoint: ENDPOINTS.COMPLEX_QUERY,
    terms,
    searchQuery: timeScope ?? "",
  };
}

/** True when this route must not touch the language model. */
export function isDeterministic(routeResult: Route): boolean {
  return routeResult.kind !== "GENERATIVE";
}