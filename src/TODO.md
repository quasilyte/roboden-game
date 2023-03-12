todo engine:
- gamepad d-pad like sticks should allow 2 directions at the same time (diagonals)
- camera
- tiled backgrounds with reversed individual tiles
- need to round the positions due to rendering issues (round them in Sprite?)
- anim like 1-2-3 played as progression 1-2-3-2 in a loop
- why animation affects Y axis?

todo:
- maybe make it possible for support drones to heal/recharge nearby colony drones
- graphics options: allow disabling some optional shaders (like scrap dissolving)
- make menu/toggle buttons transparent a bit (like other windows)
- make tutorial device-aware; if gamepad is connected, tell controller bindings; if it's a mobile device, tell about touch controls
- add turrets to a pathgrid?
- FindColonyAgent should use agents container for iteration
- add antiair missle fire effect
- show resources when base is flying
- maybe add x4 zoom scale (toggable in-game)
- consider taking a target size into account when calculating impact range
- controls rebinding
- move world generation to a background task and add a loading screen that waits for it?
- make low energy fighters fire at a slower rate
- base selector is hidden by shadow
- falling base should have damaged shader applied too
- add fullscreen option (disable/enable windowed mode)
- menu buttons focus?
- base should check landing zone before landing
- rework planner action delay (same action vs other action)
- is morale damage even viable?
- show upkeep while flying
- maybe group resources into clusters to speedup collision checking?

iskander:
- implement pause?
- make tab smooth
- esc for menu, not exit right away

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
