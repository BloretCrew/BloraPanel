import type {Page} from '@playwright/test'

// Optical-only comparison. Native CSS draws the original frame's shadow;
// no application nodes, surfaces, text or images are read or duplicated.
export const nativeCssShadow=`
.app-window{box-shadow:var(--window-shadow)!important}
.app-window.focused{box-shadow:var(--window-shadow-focus)!important}
.app-window[data-shadow-low="true"]{box-shadow:var(--window-shadow)!important}
.window-shadow-plane{display:none!important}
`
export async function installNativeCssShadow(page:Page){return page.addStyleTag({content:nativeCssShadow})}
export const nativeCssPlaneShadow=`
.window-shadow-plane[data-shadow-cache]{box-shadow:var(--window-shadow)!important;will-change:auto!important}
.app-window.focused>.window-shadow-plane[data-shadow-cache]{box-shadow:var(--window-shadow-focus)!important}
.app-window[data-shadow-low="true"]>.window-shadow-plane[data-shadow-cache]{box-shadow:var(--window-shadow)!important}
.window-shadow-plane[data-shadow-cache]>.shadow-piece{display:none!important}
`
