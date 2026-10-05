import { describe, expect, it } from "vitest";
import { route } from "./router";

/**
 * The router is the reason most requests never touch a language model, so these
 * cases assert both the classification and the `kind` that decides whether
 * generation happens at all.
 */
describe("route", () => {
  describe("event queries", () => {
    it.each([
      "show events",
      "events this week",
      "upcoming events",
      "what events are happening this week?",
    ])("routes %s to structured event search", (request) => {
      const result = route(request);

      expect(result.intent).toBe("EVENT_SEARCH");
      expect(result.kind).toBe("STRUCTURED");
      expect(result.endpoint).toBe("/events");
    });
  });

  describe("club queries", () => {
    it.each(["show clubs", "find coding clubs"])(
      "routes %s to structured club search",
      (request) => {
        const result = route(request);

        expect(result.intent).toBe("CLUB_SEARCH");
        expect(result.kind).toBe("STRUCTURED");
        expect(result.endpoint).toBe("/clubs");
      },
    );
  });

  describe("opportunity queries", () => {
    it.each(["find internships", "cloud opportunities"])(
      "routes %s to structured opportunity search",
      (request) => {
        const result = route(request);

        expect(result.intent).toBe("OPPORTUNITY_SEARCH");
        expect(result.kind).toBe("STRUCTURED");
      },
    );
  });

  it("routes resource booking to a structured route", () => {
    const result = route("book study room");

    expect(result.intent).toBe("RESOURCE_BOOKING");
    expect(result.kind).toBe("STRUCTURED");
  });

  it("routes notification queries to a structured route", () => {
    const result = route("my notifications");

    expect(result.intent).toBe("NOTIFICATION_QUERY");
    expect(result.kind).toBe("STRUCTURED");
  });

  it("routes profile queries to a structured route", () => {
    const result = route("my profile");

    expect(result.intent).toBe("PROFILE_QUERY");
    expect(result.kind).toBe("STRUCTURED");
    expect(result.endpoint).toBe("/auth/me");
  });

  describe("generation-only queries", () => {
    it.each([
      "Based on my interests, what campus opportunities should I consider?",
      "Which events are most relevant to someone interested in cloud computing?",
      "Summarize these opportunities and explain which ones are best for me.",
      "Help me understand the differences between these campus opportunities.",
      "Give me a personalized plan based on these retrieved campus resources.",
    ])("routes %s to generation", (request) => {
      const result = route(request);

      expect(result.kind).toBe("GENERATIVE");
      expect(result.intent).toBe("COMPLEX_QUERY");
    });
  });

  it("routes an unknown entity phrase to hybrid search, not generation", () => {
    const result = route("quantum photonic seminar series in the physics department");

    expect(result.kind).toBe("SEARCH");
    expect(result.intent).toBe("HYBRID_SEARCH");
  });

  it("reports low confidence for a vague question", () => {
    const result = route("something interesting");

    expect(result.confidence).toBe("low");
  });

  it("reports high confidence for an unambiguous phrase match", () => {
    // "about me" is a two-word phrase in the profile term list, which scores
    // higher than a bare keyword like "profile".
    expect(route("about me").confidence).toBe("high");
  });

  it("reports medium confidence for a single entity keyword", () => {
    expect(route("show events").confidence).toBe("medium");
  });

  it("never sends a structured request to generation", () => {
    const structured = [
      "show events",
      "show clubs",
      "find internships",
      "my notifications",
      "my profile",
      "book study room",
    ];

    for (const request of structured) {
      expect(route(request).kind).not.toBe("GENERATIVE");
    }
  });

  it("extracts searchable terms from a request", () => {
    const result = route("find coding clubs");

    expect(result.searchQuery).toContain("coding");
  });

  it("handles an empty request without throwing", () => {
    const result = route("");

    expect(result.kind).toBe("GENERATIVE");
    expect(result.terms).toEqual([]);
  });

  it("is deterministic for the same input", () => {
    expect(route("upcoming events")).toEqual(route("upcoming events"));
  });

  it("ignores casing and punctuation", () => {
    expect(route("SHOW EVENTS!").intent).toBe(route("show events").intent);
  });
});