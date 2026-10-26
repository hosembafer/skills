# Order Totals — foundation

Agents read and quote this file. They never edit it. A change is proposed in chat and a human writes it.
Owner: Dana · Status: established 2026-03-02 · SHOP-412 · PRD "Checkout totals v3" · Design frame 118-2041
Code: `libs/checkout/src/lib/feature/order-summary/`. Consumers: the invoice PDF in `libs/billing`.

## Vocabulary

- **Total** = what the customer pays, after discounts and tax. Code: `total`. Never: "amount", "price" for this.
- **Subtotal** = line items before discounts and tax. Code: `subtotal`. Copy: "Items" (PRD wording). Never: rename either side to match the other.
- **Line** = one product row in the order. Code: `line`. Never: "item", which the copy uses for the subtotal label.

## Sources

- Total, subtotal, tax, discount: `GET /orders/{id}/totals` through `@shop/data-access/orders-api`. Never: summing lines client-side; never `GET /orders/{id}` (its `total` is the pre-tax legacy field). Why: discounts are applied server-side and depend on the customer's region (SHOP-388).
- Discount codes: `POST /orders/{id}/discounts`, then refetch totals. Never: recompute the total from the code's percentage. Why: stacking rules live in the backend (SHOP-401).
- Shipping estimate: BLOCKED. No endpoint; the PRD shows it on the summary. Owner: Priya (backend).

## Access

- Viewing totals: whoever can view the order, unchanged.
- Applying a discount code: `order__update` on the order's workspace. Never: ungated, and never `order__read`. Why: support staff view orders they must not alter (SHOP-395).

## Rules

1. Totals shown are the server's last response. Never: optimistic totals after a line or discount change; show a pending state and refetch. Why: the stacking rules cannot be mirrored client-side (SHOP-401).
2. Currency formatting uses the order's currency, not the viewer's locale currency. Never: convert. Why: SHOP-377 rounding incident.
3. A removed line is gone from the summary only after the refetch confirms it. Never: filter it out locally. Why: the backend rejects removal of a line already shipped.

## Open — do not guess, ask the named person

- Does the summary show a shipping estimate before an address is entered? The PRD and the design disagree. Dana.
- Is `GET /orders/{id}` `total` going to be retired, or renamed? If retired, rule 1's refetch can use it. Priya.
