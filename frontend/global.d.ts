declare module 'next' {
  export * from 'next';
  const defaultExport: any;
  export default defaultExport;
  export type Metadata = any;
  export type NextConfig = any;
}
declare module 'next/link' { const Link: any; export = Link; export default Link; }
declare module 'next/navigation' {
  export const useRouter: any;
  export const usePathname: any;
  export const useSearchParams: any;
  export const useParams: any;
  export const redirect: any;
  export const notFound: any;
  export const permanentRedirect: any;
  export const useSelectedLayoutSegment: any;
  export const useSelectedLayoutSegments: any;
}
declare module 'next/types.js' { export type ResolvingMetadata = any; export type ResolvingViewport = any; }
