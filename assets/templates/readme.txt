# NFC Card Label Template Specifications

The Official TapTo Labels were designed for the most common NFC card size, which shares the same dimensions as a standard credit card.

Labels are meant to be placed centered on the card, with approximately 1/32” (.79mm) of exposed card at edges when applied. Through testing, we learned this accomplishes two things:
- It allows for some “wiggle room” when applying the label. If the label goes to the card edge, you have no margin for error when applying.
- It adds to the durability of the labels with handling. Labels that were applied to the edge quickly experienced rolling and peeling.

## Dimensions
Standard Credit Card size is:  
- 3.375” x 2.125”
- 85.725mm x 53.975mm

TapTo Credit Card Size Label:
- 3.313” x 2.063”
- 84.1502mm x 52.4002mm (converted from inches)
- 993.9px x 618.9px (converted from inches)

Due to the variances when importing SVG files into your design software, confirm the label size is correct. If it needs to be adjusted, use the label measurements above. For print, be sure to use a minimum of 300 dpi to retain image quality.

## Artwork Margins
Margins are measured from the edge of label.
Note: margins were rounded to nearest whole pixel for the [TapTo Designer](https://tapto-designer.netlify.app/) software, and then converted to inches / millimeters.

TapTo Horizontal margins:
- Top/Right/Bottom = 37px / .123” / 3.1242mm
- Left = 310px / 1.033” / 26.2382mm

TapTo NFC Engine margins:
- Top = 268px / .893” / 22.6822mm
- Right/Left = 37px / .123” / 3.1242mm
- Bottom = 84px / .28” / 7.112mm

TapTo Vertical margins:
- Top/Right/Left = 37px / .123” / 3.1242mm
- Bottom = 144px / .48” / 12.192mm