- allow partial updates across the api
- POS screen show price
prompt: in my pos device i have a small screen (digital numbers) it shows how much money similar to how it shows on the screen itself. is there a way to show them in my go wails app on this digital screen?

** Naming
- POS terminal
- Add images fields to expenses, purchase, products.
- Add shifts

** Today
- [ ] DB: add indexes for most used stuff. (on account ("owner", "currency"))

- [ ] Finish product creation and test completely
- [ ] Finish transfers
- [ ] UI: Finish products list and creation
- [ ] Product cannot be deleted if it exists in one of the invoices (cuz otherwise product_id which is inside it would get invalid).
- [ ] Cannot return a product unless it's in the same shift
- [ ] cannot delete sth that other stuff depend on (e.g. supplier, when product data depend on it)
- [ ] Finish orders and its queries

** Loyality System
Bronze, Silver, Gold

- You Reach a tier --> 
  Points have more value (e.g. 1.5 instead of 1)
  Points Can redeem on special edition/premium products
  Special Offer
  
*** Gold
- Can access sales before others
- Free shipping
- ?? Reserve Limited Items
  
** Stats
*** Customers
- What does he buy most (type and kind)
- What coupons did he redeem
- Daily Average income with increase & decrease
- Recurring customers percentage 
** Compare this year's profits to past year's?


** Workflows
- Close shift: update shift and cashbox money. Send email to me (maybe i should allow them first?) and to reporter as reports allow us (I can use resend or mailtrap)