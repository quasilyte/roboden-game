todo engine:
- gamepad d-pad like sticks should allow 2 directions at the same time (diagonals)
- camera
- tiled backgrounds with reversed individual tiles
- need to round the positions due to rendering issues (round them in Sprite?)
- anim like 1-2-3 played as progression 1-2-3-2 in a loop
- why animation affects Y axis?
- add Midpoint to gmath
- add distance between line and point to gmath
- add LoadGameDataRaw

computers and multi-players:
- implement defeat condition again
- "two bots" mode camera should be centered to the map
- make turrets repairable for everyone
- check whether recipe tab fits in split screen mode (spoiler: it does not)
- fix tutorial
- fix arena
- fix exit button and its notification
- BUG sometimes damage highlight stays (for colony only?)
- manual camera in replay mode
- update mode intro texts, add inf arena intro text
- cinematic camera should prioritize boss events before choosing a random player

optimizations:
- traverse creeps only once in node runner instead of twice (nodeRunner + worldState.Update)
- add elite kinds of creeps (to spawn less creeps and still hit the same budget)
- use sectors to find kamikaze explosion targets?
- do not do math.Round for spritePos in sim mode; make GetPos return real pos

todo:
- remove beam/projectile creating code duplication from drone-vs-creep
- fireoffset is duplicated in weapon and drone stats
- check if all victory conditions play a chime
- lose all evo points when moving a base?
- play error sound if can't build turret or colony
- make building construction cost more obvious and easy to balance
- make disintegrators prefer flying targets
- maybe make it possible for support drones to heal/recharge nearby colony drones
- add turrets to a pathgrid?
- FindColonyAgent should use agents container for iteration
- add antiair missle fire effect
- maybe add x4 zoom scale (toggable in-game)
- consider taking a target size into account when calculating impact range
- move world generation to a background task and add a loading screen that waits for it?
- make low energy fighters fire at a slower rate
- base selector is hidden by shadow
- falling base should have damaged shader applied too
- menu buttons focus?
- base should check landing zone before landing
- rework planner action delay (same action vs other action)
- maybe group resources into clusters to speedup collision checking?

next release:
- higher resolution
- online leaderboard
- daily run (same seed, different players, leaderboard)
- different bases (colonies)
  - bonuses and disadvantages
  - ground base (can't pass hard terrain)
  - tier 4 units for bases
- game lore
- weather
- random events
