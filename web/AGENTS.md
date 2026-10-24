# Web UI implementation and regression constraints

These rules apply to all changes under `web/`. Read them before changing routes, shared UI primitives, page containers, tabs, tables, dialogs, or other viewport-bound layouts.

## Verified layout facts

- The authenticated app shell is a fixed-height workspace: `SidebarProvider` uses `h-svh overflow-hidden`, and its route content wrapper also clips overflow (`src/components/app-shell.tsx`). Do not assume the browser document can scroll to reveal route content.
- The shared Tabs wrapper is Radix-based (`src/components/ui/tabs.tsx`). Style the state Radix actually emits (`data-state="active"`), not a guessed attribute such as `data-active`.
- The shared table already supplies its own `overflow-auto` viewport (`src/components/ui/table.tsx`); preserve that inner scroll for dense tables.

## Layout rules

1. Every flex child that must shrink within the app viewport needs `min-h-0`; horizontally constrained children also need `min-w-0`.
2. Every long page or panel inside an `overflow-hidden` ancestor must name its scroll owner explicitly. Use a bounded flex chain (`flex-1 min-h-0`) and put `overflow-y-auto` on the intended panel. `overflow-hidden` without a reachable descendant scroll area is a release blocker.
3. Choose one primary vertical scroll owner per view. For dashboards, scroll the active panel; for data lists, keep the table's own scroll viewport and keep page controls reachable. Do not make both the document and nested panel scroll accidentally.
4. With `TabsContent forceMount`, verify inactive panels remain hidden by the primitive and do not consume layout space. Do not treat “mounted” as “visible”.
5. Keep the tab's controlled `value`, `onValueChange`, URL state (when URL-backed), selected visual indicator, `aria-selected`, and displayed panel in agreement. Use the shared primitive's real data attributes for styling.

## Required regression checks for tabs and scrolling

Before calling a UI change complete, use an authenticated local test session when the page requires authentication; never bypass login. At minimum, verify in a real browser:

- Click each tab and use the keyboard (focus, arrow navigation, Enter/Space). Confirm `aria-selected`, the active indicator, visible panel content, and URL state all change together; also try browser back/forward for URL-backed tabs.
- At a desktop viewport (e.g. 1366×768), scroll from the panel top to its last section and back. Confirm no content is clipped by the app shell and the active tab/filter controls remain reachable.
- Repeat at a narrow viewport (e.g. 390×844); confirm controls wrap, the page can still be read vertically, and the table scrolls within its own viewport.
- For a scroll bug, inspect the full ancestor chain for computed height, `min-height`, `overflow-y`, `scrollHeight`, and `clientHeight`. A successful build or screenshot of only the first viewport is not scroll verification.
- Run the web typecheck, lint, and production build. If authentication or a browser is unavailable, report browser verification as not performed rather than claiming it passed.

## Regression note: historical alerts

The historical analysis panel exposed two independent mistakes: the shared Tabs trigger styles checked `data-active` while Radix exposes `data-state="active"`, so the selected-tab visual state did not track the controlled tab; and the long analysis panel sat under the fixed app shell's `overflow-hidden` layout without a bounded, vertically scrollable panel. Build and lint checks did not exercise either browser behavior. The fixes must therefore be reviewed both at the shared Tabs primitive and at the route's height/overflow chain, then verified with the checks above.

## Review checklist

- [ ] Every `overflow-hidden` in the route's ancestor chain has an intentional, reachable scroll owner (or content is proven to fit).
- [ ] Flex scroll ancestors and children have the needed `min-h-0` / `min-w-0` constraints.
- [ ] Tabs use the primitive's actual active-state attributes and expose visible selection and panel changes.
- [ ] Long-content page and table scrolling were tested at desktop and narrow viewports.
- [ ] Browser verification limitations are stated explicitly.
- [ ] Typecheck, lint, and production build passed.

## Historical alert view contract

- Keep statistical analysis and historical event details as separate tab responsibilities. The details view is a searchable event ledger; do not add analysis cards, distributions, trends, or recurrence summaries there. A result total needed by pagination may remain list metadata.
- Keep analysis URL filters separate from detail-list filters so analysis date/status selections do not silently alter event records.
- Any condition that changes historical rows must be sent through the alert-event list API and applied across the merged history set before pagination. Never implement a history filter against only the currently visible browser page.
- Historical detail filters cover summary/fingerprint, severity, rule name, datasource name, start-time range, status, and suppression. The start bound is inclusive and the end bound exclusive; convert the selected end date to the next local midnight before sending RFC3339 so sub-millisecond events at day end are not omitted.
