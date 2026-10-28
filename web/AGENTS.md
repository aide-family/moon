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

The historical analysis panel exposed two independent mistakes: the shared Tabs trigger styles checked `data-active` while Radix exposes `data-state="active"`, so the selected-tab visual state did not track the controlled tab; and the long analysis panel sat under the fixed app shell's `overflow-hidden` layout without a bounded, vertically scrollable panel. A third: `TabsContent forceMount` panels are always `present` in the primitive, so Radix renders the inactive analysis panel with `hidden={false}` and it stayed visible above the details ledger — force-mounted panels must be hidden from the controlled tab state (`className={active ? panelClasses : "hidden"}`), not left to the primitive. Build and lint checks did not exercise any of these browser behaviors. The fixes must therefore be reviewed both at the shared Tabs primitive and at the route's height/overflow chain, then verified with the checks above.

## Review checklist

- [ ] Every `overflow-hidden` in the route's ancestor chain has an intentional, reachable scroll owner (or content is proven to fit).
- [ ] Flex scroll ancestors and children have the needed `min-h-0` / `min-w-0` constraints.
- [ ] Tabs use the primitive's actual active-state attributes and expose visible selection and panel changes.
- [ ] Long-content page and table scrolling were tested at desktop and narrow viewports.
- [ ] Browser verification limitations are stated explicitly.
- [ ] Typecheck, lint, and production build passed.

## Form controls and filter panels

- Form controls keep the primitive height (`h-8`, 32px): never override `h-9`/`h-10` on `Input`, `SelectTrigger`, `Textarea`, or `DateRangePicker`. Side-by-side controls must measure equal height; verify with the rendered rects rather than by reading class names.
- List filter areas use the shared card + `FilterChips` + `FilterToggle` pattern. The toggle shares one row with the page's right-side actions (stats/batch buttons on realtime alerts, view tabs on message logs) so the header never breaks into an extra row, and the panel below collapses instead of disappearing.
- Keep visible spacing between a filter/action area and the list frame (`space-y-4` on the block container or a flex `gap`); toolbars must never sit flush against the table card.
- `FilterToggle` must point `aria-controls` at the real panel id, and the panel keeps `hidden` rather than unmounting so draft filter state survives a collapse.

## Table layout standard

All list tables share one width model, defined in `web/src/lib/table-layout.ts`:

- Main content column (name / summary / subject): no fixed width; it takes the remaining space and wraps (`break-words`). Never make the whole column `truncate`.
- Short enum columns (status / type / level): `TABLE_COLUMN_WIDTH.status|type|level` plus `whitespace-nowrap`.
- Time columns: `TABLE_COLUMN_WIDTH.time`, `TABLE_NOWRAP`, `tabular-nums`.
- Identity columns (owner / admin, name + email): `TABLE_COLUMN_WIDTH.identity`, wrap with `break-all`.
- Actions column: render its header with `TableActionsHead` (7rem, nowrap, centered) and pass `nowrap` to `RowActions`. At most two inline actions; more collapse into the "more" menu. Action buttons must never wrap to a second line.
- Table minimum width picks a tier from `TABLE_MIN_WIDTH.compact|default|wide` by column count. Horizontal overflow scrolls inside the Table/ListFrame viewport with a sticky header.
- CSS-grid "tables" (tenant list style) reuse `TABLE_GRID_MAIN_IDENTITY_STATUS_ACTIONS`. The header and every row must use the identical literal grid template — separate grid containers compute `fr`/`auto` tracks independently and drift apart otherwise.

## Historical alert view contract

- Keep statistical analysis and historical event details as separate tab responsibilities. The details view is a searchable event ledger; do not add analysis cards, distributions, trends, or recurrence summaries there. A result total needed by pagination may remain list metadata.
- Keep analysis URL filters separate from detail-list filters so analysis date/status selections do not silently alter event records.
- Any condition that changes historical rows must be sent through the alert-event list API and applied across the merged history set before pagination. Never implement a history filter against only the currently visible browser page.
- Historical detail filters cover summary/fingerprint, severity, rule name, datasource name, start-time range, status, and suppression. The start bound is inclusive (day 00:00:00) and the end bound exclusive at the selected day's 23:59:59.999999999, so the bound never drifts into the next day and fractional-second events at day end are still matched. Dates in the future are never selectable.
- Historical detail columns use the same per-user custom-column mechanism as realtime alerts, with `scope=history` (realtime uses `mine` or an alert page UID), so the two never overwrite each other. Column selection is rendered by the shared list frame action, and `ends_at` is offered only on the history side because realtime events have no end time.
- Historical date conditions use the shared dual-month `DateRangePicker` (day granularity only — never add time-of-day), capped at 30 days inclusive and never past today; the cap disables out-of-range days while picking instead of silently trimming the query. Day arithmetic lives in `src/lib/date-range.ts` and must stay DST-safe (use `dayNumber`, not raw milliseconds). The same 30-day limit is enforced server-side (`alertevent.MaxHistoryRangeDays`), so a hand-edited URL must still be rejected — do not rely on the picker alone.
- Realtime alert filters intentionally cover only summary/fingerprint and severity, in the same card layout and applied-filter chips as history. The realtime list already means "still firing", so status/suppression/time filters would duplicate its meaning; do not add them without changing that model.
- Message-log filters follow the notification API, not the alert schema: delivery logs filter by status (`SENT`/`FAILED`), task ID and created-time range; tasks filter by status and created-time range. Both send Unix seconds (`startTime` inclusive, `endTime` exclusive), cap at 30 days (server-side `notification.MaxMessageQueryDays`), and clear their filters when switching tabs because the two views do not share conditions.
- Historical analysis statistics come from six independent server-side aggregate endpoints — `GET /v1/alert-analysis/{summary,trend,status,level,fingerprints,recovery}` (local day range capped at 30 days, `timeZone`, `status`, optional `focusStart`/`focusEnd`). The analysis view must not page through `/v1/alert-events` or aggregate statistics in the browser; that path re-reads the whole merged hot+archive set once per page.
- The six blocks load sequentially in the order summary → status → level → trend → fingerprints → recovery (the next request only starts after the previous one settles), and each block owns its skeleton, error state, and retry. A filter change restarts the chain; focusing a trend bucket refetches only the five blocks outside the trend and sends the bucket as `focusStart`/`focusEnd`. Returning to an already loaded condition must not repeat requests.
- Analysis has an explicit window: without `from`/`to` in the URL the effective range is the latest 30 local days ending today, and the picker plus scope summary must show that range. Clearing analysis filters returns to it — the analysis view no longer means "all history".
- Tab data is lazy: while the history page sits on the analysis tab it must not request the details ledger (`/v1/alert-events`) or the details-side severity dictionary; switching to the details tab requests them once, and switching back does not repeat. The analysis tab lazily loads its own severity dictionary for the level distribution.
- The shared `DateRangePicker` anchors the right calendar to the current month and the left one to the previous month, both on open and after paging; the keyboard focus target must always be a visible, selectable day. The 30-day cap and future-date blocking are unchanged.

## Colors and theming

- Never write a literal color (hex/rgb/hsl) or a Tailwind palette class (`bg-emerald-500`, `text-slate-700`, …) in components. Use semantic tokens: `bg-primary` + `text-primary-foreground` for fills, `text-primary-ink` for brand-colored text/icons, `bg-primary-soft` for tinted chips, `--success/--warning/--danger/--info/--violet` (+ `-soft`/`-border`) for states, `bg-overlay` for scrims.
- Brand color is never the only state signal. Pair a color with text, an icon, weight, or position; `StatusBadge` and `StatusDot` already encode this — keep their text labels.
- 15 brands × light/dark come from three layers: `src/styles/theme/tokens.brand.css` (generated, brand scales + role colors), `semantic.light.css` / `semantic.dark.css` (role mapping, neutrals, states), and `@theme inline` in `src/index.css`. Never add `--brand-*` or `--primary*` values outside these files.
- Adding or changing a brand: edit `BRANDS` in `scripts/gen-brand-tokens.mjs`, run `npm run tokens:gen`, then `npm run tokens:check` (also verifies the `index.html` boot script's brand list). The generator asserts WCAG contrast on fill/foreground, brand text, and focus ring for both modes; a red run means the palette is not shippable.
- `--primary` is a fill, `--primary-ink` is text, `--ring` is the focus ring. Do not collapse them back into one value: yellow/teal brands cannot use one token for all three and stay legible.
