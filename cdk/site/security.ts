import {
  buildStrictCspHeader,
  type FaceCspPolicy,
  type FaceRenderResult,
} from '@theory-cloud/facetheory';

export const SITE_STYLESHEET_PATH = '/site.css';

export const strictSiteCspPolicy: FaceCspPolicy = {
  inlineScripts: false,
  inlineStyles: false,
  rawHead: false,
};

export const strictStaticPageSecurity: Pick<FaceRenderResult, 'csp' | 'headTags'> = {
  csp: strictSiteCspPolicy,
  headTags: [
    {
      type: 'link',
      attrs: {
        rel: 'stylesheet',
        href: SITE_STYLESHEET_PATH,
      },
    },
  ],
};

export const siteContentSecurityPolicy = buildStrictCspHeader();
