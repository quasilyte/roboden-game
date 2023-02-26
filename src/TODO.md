todo engine:
- gamepad d-pad like sticks should allow 2 directions at the same time (diagonals)
- camera
- tiled backgrounds with reversed individual tiles
- need to round the positions due to rendering issues (round them in Sprite?)
- anim like 1-2-3 played as progression 1-2-3-2 in a loop
- why animation affects Y axis?

todo:
- add turrets to a pathgrid?
- add damage shader for turrets
- allow workers to repair turrets?
- enemies focus 1 turret too much (see FindColonyAgent)
- FindColonyAgent should use agents container for iteration
- make choice selection window bigger or make English text fit better?
- don't send attacking units that can't attack target
- add antiair missle fire effect
- show resources when base is flying
- consider taking a target size into account when calculating impact range
- controls rebinding
- move world generation to a background task and add a loading screen that waits for it?
- make low energy fighters fire at a slower rate
- crawlers should not walk through the base
- base selector is hidden by shadow
- falling base should have damaged shader applied too
- send nearby crawlers to save the boss if it is under attack
- add fullscreen option (disable/enable windowed mode)
- menu buttons focus?
- show enemy bases on the radar
- base should check landing zone before landing
- rework planner action delay (same action vs other action)
- is morale damage even viable?
- show upkeep while flying
- maybe group resources into clusters to speedup collision checking?

iskander:
- maybe add servant drone somehow (hard mode only)
- implement pause?
- make tab smooth
- esc for menu, not exit right away
oleg:
- flamer projectile

next release:
- artifacts idea
- better how to play
+ multi-language support
- higher resolution
- more input device support (gamepad, touch screen)
- more "new game" options
- vs mode with colonies (and less creeps)
- unlockable content / achievements
- local pvp and coop
- online leaderboard
- daily run (same seed, different players, leaderboard)
- different bases (colonies)
  - bonuses and disadvantages
  - ground base (can't pass hard terrain)
  - tier 4 units for bases
- achievement conditions
  - win without faction
  - run without creep kills
  - speedrun achievement
- game modes
  - base race
  - vs colony
  - endless mode
- drone rarity
- drone pickups
- attack, inc/dec, build base, tech, build tower
- action reroll
- game lore
- weather
- random events
