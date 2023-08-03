todo engine:
- gamepad d-pad like sticks should allow 2 directions at the same time (diagonals)
- camera
- tiled backgrounds with reversed individual tiles
- tiled bg breaks if there are holes in the tilemap with percentages
- boundsrect for labels
- need to round the positions due to rendering issues (round them in Sprite?)
- anim like 1-2-3 played as progression 1-2-3-2 in a loop
- why animation affects Y axis?
- add Midpoint to gmath
- add distance between line and point to gmath
- add LoadGameDataRaw
- gmath DirectionTo works in counter-intuitive way
- an ability to "load" a transformed image (like original image + rotated hue)

new achievements:
* iron will - win 3 matches in a row without quitting
* wisp slayer - hunt down every wisp on the map

loading screen hints:
* faction bonuses

computers and multi-players:
- make turrets repairable for everyone

optimizations:
- traverse creeps only once in node runner instead of twice (nodeRunner + worldState.Update)

todo:
- remove beam/projectile creating code duplication from drone-vs-creep
- fireoffset is duplicated in weapon and drone stats
- make building construction cost more obvious and easy to balance
- consider taking a target size into account when calculating impact range
- move world generation to a background task and add a loading screen that waits for it?
