# RechargeMax UI/UX + Component Library Design Skill

> Combined skill merging **UI/UX Pro Max** (161 design rules, 67 styles, UX intelligence) with
> **Untitled UI React** (world's largest open-source React component library, Figma-parity design system).
> Stack: React + Vite + Tailwind CSS v4 + TypeScript (existing RechargeMax frontend).

---

## PART 1 — WHEN TO APPLY THIS SKILL

Use this skill whenever the task touches **how anything looks, feels, moves, or is interacted with**:

| Scenario | Examples |
|---|---|
| New page / screen | Dashboard, landing page, prize history, spin wheel |
| New component | Buttons, modals, cards, forms, toasts, nav |
| Style / color / font decisions | "What style fits a fintech/gamified rewards app?" |
| UI review or audit | "Review this page for UX issues" |
| Fix a visual bug | "Layout shifts on load", "Button hover broken" |
| Responsive / mobile pass | "Make this work on Android Chrome" |
| Accessibility pass | "Check keyboard nav and contrast" |
| Dark mode | "Add dark mode support" |
| Animation / micro-interactions | "Add smooth transitions" |
| Charts / data viz | "Add prize stats chart" |

**Skip for:** pure backend logic, API design, DB migrations, DevOps, scripts with no visual output.

---

## PART 2 — DESIGN RULES BY PRIORITY (161 rules, 10 categories)

Follow priority 1 → 10 when making design decisions.

### Priority 1 — Accessibility (CRITICAL)

- Minimum contrast **4.5:1** for normal text; 3:1 for large text (≥18px bold, ≥24px regular)
- **Visible focus rings** on all interactive elements — 2px outline, never `outline: none`
- Alt text on all meaningful images; `aria-label` on icon-only buttons
- Full keyboard navigation: Tab order matches visual order
- `role` + `aria-*` on custom components (modals, dropdowns, spin wheel)
- Never convey information by color alone — always add icon/text
- Respect `prefers-reduced-motion` — disable/reduce animations when set
- Semantic HTML: `<button>` for actions, `<a>` for links, `<h1–h6>` in order

### Priority 2 — Touch & Interaction (CRITICAL)

- Min touch target: **44×44px** (Apple HIG); 48×48dp (Material)
- Min 8px gap between interactive elements
- Disable button during async operations; show spinner or progress
- `cursor: pointer` on all clickable elements
- `touch-action: manipulation` to remove 300ms tap delay on mobile
- Visual press feedback within 100ms of tap
- Never rely on hover-only for primary interactions

### Priority 3 — Performance (HIGH)

- WebP/AVIF images; `loading="lazy"` for below-fold assets
- Declare `width`/`height` or `aspect-ratio` to prevent CLS
- `font-display: swap` for web fonts
- Route-level code splitting (React.lazy / dynamic imports)
- Virtualize lists with 50+ items
- Skeleton screens for >300ms loads, not a spinner
- Debounce high-frequency events (scroll, resize, search input)

### Priority 4 — Style Selection (HIGH) — **RechargeMax = Vibrant Fintech + Gamification**

RechargeMax is a **gamified mobile recharge rewards platform** targeting Nigerian consumers (MTN).
Recommended style blend:

| Layer | Style | Rationale |
|---|---|---|
| Base | **Modern Fintech** (clean cards, trustworthy) | Financial transactions need trust signals |
| Gamification | **Vibrant & Block-based** (bold colors, energetic) | Spin wheel, prizes, excitement |
| Micro-accents | **Glassmorphism** on modals/overlays | Premium feel on prize win moments |
| Dark mode | Optional OLED-optimized | Battery savings on Nigerian Android devices |

**Brand palette:**
```css
--color-brand-primary: #7C3AED;   /* Purple — primary actions, spin wheel */
--color-brand-secondary: #F59E0B; /* Amber/Gold — prizes, rewards, CTA highlights */
--color-success: #10B981;          /* Green — successful recharge, claim success */
--color-error: #EF4444;            /* Red — errors, destructive actions */
--color-neutral-900: #111827;      /* Text primary */
--color-neutral-500: #6B7280;      /* Text secondary */
--color-surface: #FFFFFF;          /* Card / modal background */
--color-bg: #F9FAFB;               /* Page background */
```

- Use SVG icons only (Untitled UI Icons, Lucide) — no emoji as icons
- One icon library consistently (Untitled UI Icons `@untitledui/icons`)
- Each screen has **one primary CTA** — secondary actions visually subordinate
- Consistent elevation scale: cards → modals → toasts → tooltips

### Priority 5 — Layout & Responsive (HIGH)

- **Mobile-first** — design for 375px then scale up
- Breakpoints: 375 / 768 / 1024 / 1440
- `min-h-dvh` not `100vh` on mobile
- Max container width: `max-w-7xl` (1280px)
- 4pt/8px spacing increments (4, 8, 12, 16, 24, 32, 48, 64, 96)
- No horizontal scroll ever
- Fixed navbar reserves padding for underlying content
- Consistent z-index scale:

```
z-10   Dropdowns, tooltips
z-20   Sticky headers
z-40   Modals, drawers
z-50   Toasts, notifications
z-100  Spin wheel overlay
```

### Priority 6 — Typography & Color (MEDIUM)

- Base body: **16px / 1.5 line-height** (avoids iOS auto-zoom)
- Type scale: 12 / 14 / 16 / 18 / 20 / 24 / 30 / 36 / 48px
- Font stack: `Inter` (UI text) + `Plus Jakarta Sans` (headings/display)
- **Always use semantic color tokens** — never raw hex in components:

```css
/* ✅ Correct */
className="text-brand-primary bg-surface border-primary"

/* ❌ Wrong */
className="text-purple-700 bg-white border-gray-200"
```

Semantic color classes to use:
- `text-primary` — main body text
- `text-secondary` — supporting/muted text
- `text-brand-primary` / `text-brand-secondary` — brand colors
- `text-error-primary` — error states
- `bg-primary` / `bg-secondary` — surface backgrounds
- `bg-brand-primary` — brand-colored backgrounds
- `border-primary` / `border-brand` — borders

### Priority 7 — Animation (MEDIUM)

- Duration: **150–300ms** for micro-interactions; ≤400ms for complex
- Only animate `transform` and `opacity` — never `width`, `height`, `top`, `left`
- Ease-out for entering, ease-in for exiting
- Spring/physics-based curves for prize reveals and spin results
- Spin wheel: 3–5s physics spin, easing out, then prize reveal with scale+glow
- `prefers-reduced-motion`: disable spin animation, show result instantly
- Exit animations 60–70% the duration of enter animations

```css
/* Standard micro-interaction */
transition: all 150ms ease-out;

/* Prize reveal */
transition: transform 300ms cubic-bezier(0.34, 1.56, 0.64, 1);

/* Spin wheel deceleration */
transition: transform 4000ms cubic-bezier(0.17, 0.67, 0.12, 0.99);
```

### Priority 8 — Forms & Feedback (MEDIUM)

- Every input has a **visible label** — never placeholder-only
- Errors appear **below the relevant field**, not only at top
- Loading → success/error state on every form submission
- Auto-dismiss toasts in **3–5 seconds**; manual dismiss also available
- Confirm before destructive actions (claim prize, logout)
- Validate on **blur**, not on every keystroke
- `aria-live="polite"` on toast container for screen reader announcements
- Phone number input: `type="tel"` triggers numeric keyboard on mobile

### Priority 9 — Navigation (HIGH)

- Bottom nav on mobile (≤768px): max 5 items, icon + label always
- Sidebar on desktop (≥1024px)
- Current page highlighted in nav (color + weight indicator)
- Back behavior predictable — preserve scroll position
- Modals have clear close affordance (X button + Esc key + backdrop click)
- **RechargeMax nav structure:**
  - Mobile bottom: Home | Recharge | Spin | Prizes | Account
  - Desktop sidebar: Dashboard, Recharge, Spin Wheel, Prize History, Draw History, Account

### Priority 10 — Charts & Data (LOW)

- Match chart type to data: trend → line, comparison → bar, proportion → donut
- Accessible color palette (colorblind-safe)
- Always provide tooltip on hover/tap with exact values
- Recharts (already in project) or Chart.js — consistent across all charts
- Empty state: "No data yet" + guidance, never a blank chart frame

---

## PART 3 — UNTITLED UI REACT COMPONENT LIBRARY

### Stack & Installation

The existing RechargeMax frontend runs **React + Vite + Tailwind CSS v4 + TypeScript**.
Untitled UI React components are MIT-licensed and can be copied directly.

**Install dependencies:**
```bash
npm install react-aria-components react-aria tailwind-merge tailwindcss-animate motion
npm install @untitledui/icons @untitledui/file-icons
npm install tailwindcss-react-aria-components
```

### Critical Conventions

**1. React Aria import prefix (MANDATORY):**
```typescript
// ✅ Always prefix with Aria*
import { Button as AriaButton, TextField as AriaTextField } from "react-aria-components";

// ❌ Never import without prefix — causes naming conflicts
import { Button } from "react-aria-components";
```

**2. File naming (kebab-case everywhere):**
```
✅  prize-history.tsx  |  user-dashboard.tsx  |  spin-wheel.tsx
❌  PrizeHistory.tsx   |  UserDashboard.tsx   |  SpinWheel.tsx
```

**3. MUST use semantic color classes — never raw Tailwind color classes:**
```typescript
// ✅ Correct — semantic
<p className="text-primary">Total Points</p>
<div className="bg-brand-primary text-white">Claim Now</div>

// ❌ Wrong — raw
<p className="text-gray-900">Total Points</p>
<div className="bg-purple-600 text-white">Claim Now</div>
```

### Component Reference

#### Button
```typescript
import { Button } from "@/components/base/buttons/button";

// Props
size:  "sm" | "md" | "lg" | "xl"         // default: "sm"
color: "primary" | "secondary" | "tertiary"
     | "link-gray" | "link-color"
     | "primary-destructive" | "secondary-destructive"
     | "tertiary-destructive" | "link-destructive"
iconLeading:          FC | ReactNode      // icon before text
iconTrailing:         FC | ReactNode      // icon after text
isDisabled:           boolean
isLoading:            boolean             // shows spinner
showTextWhileLoading: boolean

// Usage
<Button size="md" color="primary">Recharge Now</Button>
<Button size="md" color="primary" iconLeading={CreditCard02} isLoading>Processing...</Button>
<Button size="sm" color="primary-destructive" iconLeading={Trash02}>Cancel Claim</Button>
<Button color="link-gray" href="/prizes">View All Prizes</Button>
```

#### Input
```typescript
import { Input } from "@/components/base/input/input";
import { InputGroup } from "@/components/base/input/input-group";

// Props
size:       "sm" | "md"
label:      string                  // always provide
placeholder:string
hint:       string                  // helper text below
tooltip:    string                  // help icon tooltip
icon:       FC                      // leading icon
isRequired: boolean                 // adds asterisk
isDisabled: boolean
isInvalid:  boolean                 // red error state

// Usage — Phone recharge form
<Input
  label="Phone Number"
  placeholder="08XXXXXXXXX"
  icon={Phone01}
  isRequired
  type="tel"
  hint="Enter 11-digit MTN number"
/>

// With inline action
<InputGroup label="Amount (₦)" trailingAddon={<Button size="sm">Recharge</Button>}>
  <InputBase placeholder="1000" type="number" />
</InputGroup>
```

#### Select
```typescript
import { Select } from "@/components/base/select/select";
import { MultiSelect } from "@/components/base/select/multi-select";

// Basic
<Select label="Network" placeholder="Select network" items={networks}>
  {(item) => (
    <Select.Item id={item.id} supportingText={item.description}>
      {item.name}
    </Select.Item>
  )}
</Select>

// With search (ComboBox)
<Select.ComboBox label="Prize Type" placeholder="Search..." items={prizeTypes}>
  {(item) => <Select.Item id={item.id}>{item.name}</Select.Item>}
</Select.ComboBox>
```

#### Badge
```typescript
import { Badge } from "@/components/base/badges/badges";

// type: "pill-color" | "color" | "modern"
// color: "gray" | "brand" | "error" | "warning" | "success"
// size: "sm" | "md" | "lg"

<Badge type="pill-color" color="success" size="sm">Claimed</Badge>
<Badge type="pill-color" color="warning" size="sm">Pending</Badge>
<Badge type="pill-color" color="error" size="sm">Rejected</Badge>
<Badge type="pill-color" color="brand" size="sm">Pending Review</Badge>
<Badge type="color" color="gray" size="sm">Approved</Badge>
```

#### Avatar
```typescript
import { Avatar } from "@/components/base/avatar/avatar";
import { AvatarLabelGroup } from "@/components/base/avatar/avatar-label-group";

// Props
size:            "xs" | "sm" | "md" | "lg" | "xl" | "2xl"
src:             string           // image URL
alt:             string
initials:        string           // fallback "AO"
status:          "online" | "offline" | "away" | "busy"
verified:        boolean

<Avatar size="md" initials="JD" status="online" />
<AvatarLabelGroup
  size="md"
  name="John Doe"
  supportingText="08011111111"
  src={user.avatarUrl}
/>
```

#### FeaturedIcon
```typescript
import { FeaturedIcon } from "@/components/foundations/featured-icon/featured-icon";

// Great for empty states, prize reveal, success screens
// theme: "light" | "gradient" | "dark" | "modern" | "modern-neue" | "outline"
// color: "brand" | "gray" | "error" | "warning" | "success"

// Prize win celebration
<FeaturedIcon icon={Gift01} size="xl" color="brand" theme="gradient" />

// Empty prize history
<FeaturedIcon icon={Trophy01} size="lg" color="gray" theme="light" />
```

#### Modal
```typescript
import { Modal } from "@/components/application/modals/modal";

// Use for: Spin Wheel, Prize Claim, Confirmation dialogs
// Always: ESC to close, backdrop click to close, X button visible
<Modal>
  <Modal.Trigger asChild>
    <Button>Spin Now</Button>
  </Modal.Trigger>
  <Modal.Content>
    <Modal.Header title="Spin the Wheel!" description="You have 2 spins remaining." />
    <Modal.Body>
      {/* SpinWheelCanvas component */}
    </Modal.Body>
    <Modal.Footer>
      <Button color="secondary" onPress={onClose}>Cancel</Button>
      <Button color="primary" onPress={onSpin} isLoading={isSpinning}>
        Spin!
      </Button>
    </Modal.Footer>
  </Modal.Content>
</Modal>
```

#### Tabs
```typescript
import { Tabs } from "@/components/application/tabs/tabs";

// Use for: Dashboard sections, Prize history filters
<Tabs defaultSelectedKey="airtime">
  <Tabs.List>
    <Tabs.Tab id="airtime">Airtime</Tabs.Tab>
    <Tabs.Tab id="data">Data</Tabs.Tab>
    <Tabs.Tab id="cash">Cash</Tabs.Tab>
  </Tabs.List>
  <Tabs.Panel id="airtime">{/* airtime prizes */}</Tabs.Panel>
  <Tabs.Panel id="data">{/* data prizes */}</Tabs.Panel>
  <Tabs.Panel id="cash">{/* cash prizes */}</Tabs.Panel>
</Tabs>
```

#### Table
```typescript
import { Table } from "@/components/application/table/table";

// Use for: Recharge history, API transactions, Draw history
<Table aria-label="Recharge History">
  <Table.Header>
    <Table.Column isRowHeader>Date</Table.Column>
    <Table.Column>Amount</Table.Column>
    <Table.Column>Network</Table.Column>
    <Table.Column>Status</Table.Column>
  </Table.Header>
  <Table.Body items={recharges}>
    {(row) => (
      <Table.Row id={row.id}>
        <Table.Cell>{formatDate(row.date)}</Table.Cell>
        <Table.Cell>₦{formatAmount(row.amount)}</Table.Cell>
        <Table.Cell>{row.network}</Table.Cell>
        <Table.Cell>
          <Badge color={row.status === "success" ? "success" : "error"}>
            {row.status}
          </Badge>
        </Table.Cell>
      </Table.Row>
    )}
  </Table.Body>
</Table>
```

#### Pagination
```typescript
import { Pagination } from "@/components/application/pagination/pagination";

<Pagination
  totalPages={totalPages}
  currentPage={page}
  onPageChange={setPage}
/>
```

#### Empty State
```typescript
import { EmptyState } from "@/components/application/empty-state/empty-state";

<EmptyState
  icon={Trophy01}
  title="No prizes yet"
  description="Spin the wheel to win airtime, data, and cash prizes."
  action={<Button color="primary">Spin Now</Button>}
/>
```

#### Loading Indicator
```typescript
import { LoadingIndicator } from "@/components/application/loading-indicator/loading-indicator";

// size: "sm" | "md" | "lg"
// For page-level loading
<LoadingIndicator size="lg" />
```

### Utility Functions

```typescript
import { cx } from "@/utils/cx";
import { sortCx } from "@/utils/cx";

// cx: merge class names (like clsx + tailwind-merge)
className={cx("base-class", isActive && "active-class", className)}

// sortCx: organized style objects
const styles = sortCx({
  common: { root: "flex items-center gap-2" },
  sizes:  { sm: { root: "h-8 px-3 text-sm" }, md: { root: "h-10 px-4 text-base" } },
  colors: { primary: { root: "bg-brand-primary text-white" } },
});
```

### Icon Usage
```typescript
import { Gift01, Trophy01, Phone01, CreditCard02, ChevronDown, X } from "@untitledui/icons";

// In component props (preferred — tree-shakeable)
<Button iconLeading={Gift01}>Claim Prize</Button>

// Standalone
<Gift01 className="size-5 text-brand-primary" aria-hidden="true" />

// With stroke width
<Trophy01 className="size-6 text-fg-brand-primary" strokeWidth={1.5} />

// Sizing guide
// size-4 = 16px  (inline, badge)
// size-5 = 20px  (button icon, nav item)
// size-6 = 24px  (feature icon, header)
// size-8 = 32px  (empty state, illustration)
```

### Brand Theme Configuration

Add to `src/styles/theme.css`:
```css
:root {
  /* RechargeMax Brand Colors — Purple + Gold rewards palette */
  --color-brand-25:  #faf5ff;
  --color-brand-50:  #f3e8ff;
  --color-brand-100: #e9d5ff;
  --color-brand-200: #d8b4fe;
  --color-brand-300: #c084fc;
  --color-brand-400: #a855f7;
  --color-brand-500: #9333ea;
  --color-brand-600: #7c3aed;  /* Primary interactive */
  --color-brand-700: #6d28d9;
  --color-brand-800: #5b21b6;
  --color-brand-900: #4c1d95;
  --color-brand-950: #2e1065;

  /* Reward / Prize Gold */
  --color-reward-400: #fbbf24;
  --color-reward-500: #f59e0b;
  --color-reward-600: #d97706;
}
```

---

## PART 4 — DESIGN TOKENS: THREE-LAYER ARCHITECTURE

```
Primitive (raw values)  →  Semantic (purpose aliases)  →  Component (component-specific)
```

```css
/* Layer 1: Primitive */
--color-purple-600: #7c3aed;
--color-amber-500:  #f59e0b;
--space-4: 1rem;

/* Layer 2: Semantic */
--color-primary:        var(--color-purple-600);
--color-cta-highlight:  var(--color-amber-500);
--spacing-component:    var(--space-4);

/* Layer 3: Component */
--button-bg-primary:    var(--color-primary);
--prize-card-highlight: var(--color-cta-highlight);
--card-padding:         var(--spacing-component);
```

**Rule:** Components import semantic tokens only — never primitive tokens directly.

---

## PART 5 — RECHARGEMAX-SPECIFIC COMPONENT PATTERNS

### Prize Card (PENDING state)
```tsx
<div className="rounded-xl border border-primary bg-surface p-4 space-y-3">
  <div className="flex items-start justify-between">
    <div className="space-y-1">
      <p className="text-sm font-semibold text-primary">{prize.prizeName}</p>
      <p className="text-xs text-secondary">Won on {formatDate(prize.wonAt)}</p>
      {prize.fulfillmentMode && (
        <p className="text-xs text-tertiary">Mode: {prize.fulfillmentMode}</p>
      )}
    </div>
    <div className="text-right space-y-1">
      <p className="text-sm font-bold text-primary">
        ₦{prize.prizeValue.toLocaleString()}
      </p>
      <Badge type="pill-color" color="warning" size="sm">PENDING</Badge>
    </div>
  </div>
  <Button size="sm" color="primary" className="w-full" onPress={() => onClaim(prize.id)}>
    <Gift01 className="size-4" /> Claim Now
  </Button>
</div>
```

### Prize Card (PENDING_ADMIN_REVIEW state)
```tsx
<div className="rounded-xl border border-warning bg-warning-subtle p-4 space-y-2">
  <div className="flex items-start justify-between">
    <p className="text-sm font-semibold text-primary">{prize.prizeName}</p>
    <Badge type="pill-color" color="warning" size="sm">Under Review</Badge>
  </div>
  <p className="text-xs text-secondary">
    Your claim is being reviewed. We'll notify you once processed.
  </p>
  {prize.claimReference && (
    <p className="text-xs font-mono text-tertiary">Ref: {prize.claimReference}</p>
  )}
</div>
```

### Spin Wheel Toast (instead of auto-popup)
```tsx
// Show when user is eligible but hasn't explicitly clicked Spin
toast({
  title: "🎡 You have spins available!",
  description: `${spinsRemaining} spin${spinsRemaining > 1 ? "s" : ""} ready to use`,
  action: <Button size="sm" color="primary" onPress={openSpinWheel}>Spin Now</Button>,
  duration: 5000,
});
```

### Empty States
```tsx
// No prizes yet
<EmptyState
  icon={Trophy01}
  title="No prizes won yet"
  description="Recharge ₦1,000+ to earn spins and win airtime, data, and cash."
  action={<Button color="primary" iconLeading={CreditCard02}>Recharge Now</Button>}
/>

// Spins exhausted
<EmptyState
  icon={RefreshCw01}
  title="All spins used for today"
  description={`Recharge ₦${amountNeeded.toLocaleString()} more to unlock ${nextTierSpins} spins.`}
  action={<Button color="primary">Recharge to Unlock</Button>}
/>
```

---

## PART 6 — STYLE RULES QUICK CHECKLIST

Before committing any UI change, verify:

```
ACCESSIBILITY
☐ All interactive elements have visible focus ring
☐ Contrast ≥ 4.5:1 for text, ≥ 3:1 for large text
☐ Icon-only buttons have aria-label
☐ Custom components have role + aria-* attributes
☐ prefers-reduced-motion respected

TOKENS & CLASSES
☐ Zero raw hex values in components
☐ Zero raw Tailwind color classes (text-gray-*, bg-purple-*) in components
☐ All colors use semantic token classes (text-primary, bg-brand-primary etc.)
☐ Spacing uses 4pt/8px increments

MOBILE
☐ Touch targets ≥ 44×44px
☐ No horizontal scroll at 375px
☐ Input type="tel" for phone numbers
☐ type="number" for amounts

COMPONENTS (Untitled UI)
☐ React Aria imports prefixed with Aria*
☐ Files named in kebab-case
☐ cx() used for conditional classes
☐ All inputs have visible label prop (not just placeholder)
☐ Loading state on all async buttons (isLoading prop)

ANIMATIONS
☐ Only transform/opacity animated
☐ Duration 150–300ms for micro-interactions
☐ prefers-reduced-motion: disable non-essential animation
☐ Spin wheel respects reduced-motion

FORMS
☐ Errors shown below field, not just at page top
☐ Submit button disabled + spinner during request
☐ aria-live="polite" on error/success regions
```

---

## PART 7 — ANTI-PATTERNS (NEVER DO)

| Anti-Pattern | Correct Approach |
|---|---|
| Raw hex in components | Use semantic CSS token class |
| Raw Tailwind colors (text-gray-900) | Use semantic classes (text-primary) |
| Icon-only button with no aria-label | `aria-label="Claim prize"` |
| `outline: none` on focus | Visible focus ring always |
| Auto-opening spin wheel modal on recharge | Show toast with "Spin Now" button instead |
| Placeholder as label | Always provide `label` prop |
| `position: fixed` elements hiding content | Reserve padding for fixed elements |
| `overflow: hidden` clipping focus rings | Use `overflow: clip` or add padding |
| Emoji as icons | SVG from `@untitledui/icons` |
| Multiple primary CTAs per screen | One primary, rest secondary/tertiary |
| Hard-coded `₦20,000` for a ₦200 prize | Always divide kobo by 100 server-side |
| Raw hex in CSS: `color: #7c3aed` | `color: var(--color-brand-600)` |

---

## PART 8 — RECHARGEMAX PAGES REFERENCE

| Page | Key Components | Primary CTA |
|---|---|---|
| Home/Dashboard | Stats cards, tier badge, recent activity | "Recharge Now" |
| Recharge | Phone input, network select, amount input | "Recharge" |
| Spin Wheel | Modal + canvas wheel, spin count badge | "Spin!" |
| Prize History | Prize cards (PENDING/REVIEW/CLAIMED), tabs | "Claim Now" |
| Draw History | Table + pagination | — |
| Account | Avatar, profile form, tier progress | "Save Changes" |
| Admin Dashboard | Stats, winner table, draw controls | "Run Draw" |

---

*Sources: UI/UX Pro Max Skill (MIT) — github.com/nextlevelbuilder/ui-ux-pro-max-skill*
*         Untitled UI React (MIT) — github.com/untitleduico/untitledui-nextjs-starter-kit*
