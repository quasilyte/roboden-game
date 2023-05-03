todo engine:
- gamepad d-pad like sticks should allow 2 directions at the same time (diagonals)
- camera
- tiled backgrounds with reversed individual tiles
- need to round the positions due to rendering issues (round them in Sprite?)
- anim like 1-2-3 played as progression 1-2-3-2 in a loop
- why animation affects Y axis?
- add Midpoint to gmath
- add LoadGameDataRaw

optimizations:
- make processSupport faster, have a flag hasSupport in drone stats
- make creep.TargetKind work faster (don't use IsFlying there)
- use different movetowards (dir vector?)
- add elite kinds of creeps (to spawn less creeps and still hit the same budget)
- don't create projectile sprites
- don't create projectile explosion effects when simulating
- don't create projectile trail effects when simulating
- maybe use a reusable projectileNode pool
- optimize "find targets around point" code; don't scan everyone
- do not do math.Round for spritePos in sim mode; make GetPos return real pos
bugs:
- victoryCheckDelay in nodeRunner is unused?

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
