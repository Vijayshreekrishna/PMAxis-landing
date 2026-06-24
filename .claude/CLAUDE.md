# Claude Skills

---

## 1. Frontend Design

Approach this as the design lead at a small studio known for giving every client a visual identity that could not be mistaken for anyone else's. Make deliberate, opinionated choices about palette, typography, and layout that are specific to this brief, and take one real aesthetic risk you can justify.

### Ground it in the subject
If the brief does not pin down what the product or subject is, pin it yourself before designing: name one concrete subject, its audience, and the page's single job, and state your choice. The subject's own world — its materials, instruments, artifacts, and vernacular — is where distinctive choices come from.

### Design principles
- The hero is a thesis. Open with the most characteristic thing in the subject's world.
- Typography carries the personality of the page. Pair display and body faces deliberately. Make the type treatment itself a memorable part of the design.
- Structure is information. Structural devices should encode something true about the content, not decorate it.
- Leverage motion deliberately. An orchestrated moment usually lands harder than scattered effects.
- Match complexity to the vision. Maximalist directions need elaborate execution; minimal directions need precision.

### Process
Work in two passes:
1. **Brainstorm a design plan**: compact token system with color (4–6 named hex values), type (2+ roles), layout (ASCII wireframes), and signature (the single unique element).
2. **Review against brief**: if any part reads like a generic default, revise it before writing code.

### Restraint
Spend your boldness in one place. Let the signature element be the one memorable thing, keep everything else quiet and disciplined. Build responsively, with visible keyboard focus and reduced motion respected.

### Writing in design
Words appear in a design for one reason: to make it easier to understand. Write from the end user's side of the screen. Use active voice. Keep the register conversational: plain verbs, sentence case, no filler.

---

## 2. Minimalist UI (Premium Utilitarian Style)

### Absolute rules — never do these
- Do NOT use Inter, Roboto, or Open Sans
- Do NOT use Lucide, Feather, or standard Heroicons
- Do NOT use heavy drop shadows (`shadow-md`, `shadow-lg`, `shadow-xl`)
- Do NOT use primary-colored backgrounds for large sections
- Do NOT use gradients, neon colors, or 3D glassmorphism
- Do NOT use `rounded-full` for large containers or buttons
- Do NOT use emojis anywhere — use proper icons or SVG primitives
- Do NOT use generic placeholder names like "John Doe" or "Lorem Ipsum"
- Do NOT use AI clichés: "Elevate", "Seamless", "Unleash", "Next-Gen", "Game-changer"

### Typography
- **Body/UI/Buttons**: `'SF Pro Display', 'Geist Sans', 'Helvetica Neue', 'Switzer', sans-serif`
- **Hero headings**: `'Lyon Text', 'Newsreader', 'Playfair Display', 'Instrument Serif', serif` — tight tracking (`-0.02em` to `-0.04em`), tight line-height (`1.1`)
- **Code/Meta**: `'Geist Mono', 'SF Mono', 'JetBrains Mono', monospace`
- Body text: off-black `#111111` or `#2F3437`, never pure black. Line-height `1.6`. Secondary: `#787774`

### Color palette
- Canvas: `#FFFFFF` or `#F7F6F3` / `#FBFBFA`
- Cards: `#FFFFFF` or `#F9F9F8`
- Borders/Dividers: `#EAEAEA` or `rgba(0,0,0,0.06)`
- Accents (muted pastels only):
  - Pale Red: `#FDEBEC` (text `#9F2F2D`)
  - Pale Blue: `#E1F3FE` (text `#1F6C9F`)
  - Pale Green: `#EDF3EC` (text `#346538`)
  - Pale Yellow: `#FBF3DB` (text `#956400`)

### Components
- **Cards**: `border: 1px solid #EAEAEA`, `border-radius: 8px–12px`, padding `24px–40px`
- **Buttons**: `background: #111111`, `color: #FFFFFF`, `border-radius: 4px–6px`, no shadow. Hover: `#333333` or `scale(0.98)`
- **Tags**: pill-shaped, `text-xs`, uppercase, wide tracking, muted pastel bg
- **Accordions**: no container boxes, only `border-bottom: 1px solid #EAEAEA`, `+`/`-` toggle icon
- **Kbd shortcuts**: `<kbd>` with `border: 1px solid #EAEAEA`, `border-radius: 4px`, `background: #F7F6F3`, monospace font

### Motion
- Scroll entry: `translateY(12px)` + `opacity: 0` → resolved over `600ms` with `cubic-bezier(0.16, 1, 0.3, 1)`. Use `IntersectionObserver`.
- Hover: cards lift with `box-shadow: 0 2px 8px rgba(0,0,0,0.04)` over `200ms`
- Staggered reveals: `animation-delay: calc(var(--index) * 80ms)`
- Animate via `transform` and `opacity` only — no layout-triggering properties

### Execution order
1. Establish macro-whitespace first (`py-24` or `py-32`)
2. Constrain content width to `max-w-4xl` or `max-w-5xl`
3. Apply typographic hierarchy and monochromatic color variables
4. Enforce `1px solid #EAEAEA` on all cards, dividers, borders
5. Add scroll-entry animations to all major content blocks

---

## 3. Brand Guidelines (Anthropic Style)

### Colors
**Main:**
- Dark: `#141413` — primary text and dark backgrounds
- Light: `#faf9f5` — light backgrounds and text on dark
- Mid Gray: `#b0aea5` — secondary elements
- Light Gray: `#e8e6dc` — subtle backgrounds

**Accents:**
- Orange: `#d97757` — primary accent
- Blue: `#6a9bcc` — secondary accent
- Green: `#788c5d` — tertiary accent

### Typography
- **Headings (24pt+)**: Poppins (Arial fallback)
- **Body text**: Lora (Georgia fallback)

### Application rules
- Apply Poppins to headings, Lora to body text
- Use accent colors (orange → blue → green) for non-text shapes and highlights
- Smart color selection based on background (dark bg = light text, light bg = dark text)
- Preserve text hierarchy and formatting at all times

---

## 4. Doc Co-Authoring Workflow

When asked to write documentation, proposals, technical specs, decision docs, or similar structured content, follow this three-stage workflow.

### Stage 1: Context Gathering
**Goal:** Close the gap between what the user knows and what Claude knows.

Ask these upfront:
1. What type of document is this? (spec, decision doc, proposal, etc.)
2. Who's the primary audience?
3. What's the desired impact when someone reads this?
4. Is there a template or specific format to follow?
5. Any other constraints or context?

Then encourage an info dump — background, team context, alternatives considered, timeline pressures, stakeholder concerns. Ask clarifying questions (5–10) after the dump. Exit Stage 1 when edge cases and trade-offs can be asked about without needing basics explained.

### Stage 2: Refinement & Structure
**Goal:** Build the document section by section.

For each section:
1. Ask 5–10 clarifying questions about what to include
2. Brainstorm 5–20 options for content
3. User curates: keep/remove/combine
4. Draft the section
5. Refine through surgical edits (use `str_replace`, never reprint the whole doc)

Start with the section that has the most unknowns. Create a scaffold with all section headers first. After 3 iterations with no substantial changes, ask if anything can be removed.

### Stage 3: Reader Testing
**Goal:** Test the doc with a fresh Claude (no context) to catch blind spots.

1. Predict 5–10 questions a reader would realistically ask
2. Test those questions with a fresh Claude instance (open new chat, paste doc, ask the questions)
3. Also ask: "What in this doc might be ambiguous?", "What knowledge does this assume?", "Are there contradictions?"
4. Fix any gaps found, loop back to refinement if needed

**Doc is ready** when Reader Claude consistently answers questions correctly.

### Final tips
- Do a final read-through yourself — you own this document
- Double-check facts, links, and technical details
- Consider linking this conversation in an appendix
- Update the doc as feedback comes in from real readers