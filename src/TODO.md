todo engine:
- gamepad d-pad like sticks should allow 2 directions at the same time (diagonals)
- camera
- tiled backgrounds with reversed individual tiles
- need to round the positions due to rendering issues (round them in Sprite?)
- anim like 1-2-3 played as progression 1-2-3-2 in a loop
- why animation affects Y axis?

todo:
- make disintegrators prefer flying targets
- maybe make it possible for support drones to heal/recharge nearby colony drones
- graphics options: allow disabling some optional shaders (like scrap dissolving)
- make menu/toggle buttons transparent a bit (like other windows)
- make tutorial device-aware; if gamepad is connected, tell controller bindings; if it's a mobile device, tell about touch controls
- add turrets to a pathgrid?
- FindColonyAgent should use agents container for iteration
- add antiair missle fire effect
- maybe add x4 zoom scale (toggable in-game)
- consider taking a target size into account when calculating impact range
- controls rebinding
- move world generation to a background task and add a loading screen that waits for it?
- make low energy fighters fire at a slower rate
- base selector is hidden by shadow
- falling base should have damaged shader applied too
- menu buttons focus?
- base should check landing zone before landing
- rework planner action delay (same action vs other action)
- show upkeep while flying
- maybe group resources into clusters to speedup collision checking?

iskander:
- implement pause?
- make tab smooth
- esc for menu, not exit right away

next release:
- artifacts idea
+ better how to play
+ multi-language support
- higher resolution
+ more input device support (gamepad, touch screen)
+ more "new game" options
- online leaderboard
- daily run (same seed, different players, leaderboard)
- different bases (colonies)
  - bonuses and disadvantages
  - ground base (can't pass hard terrain)
  - tier 4 units for bases
- game modes
  - arena
+ towers
- game lore
- weather
- random events

tech list (generic):
- technology seal: disable tech options from the 5th action
- increased colony movement speed (+10%), 2 levels
- increased colony movement range (+15%), 2 levels
- extra drone hp (+10%), 2 levels
- increased turret hp (+40%)
- passive hp regen for buildings
- increased drone energy-damage resist (50%)
- increased drone energy regen (+10%), 2 levels
- faster (and cheaper) building construction (+33%)
- max resource capacity increase for colonies (+20%)
- increased colony drone count limit (+10%)
- faster research speed (+15%)
- higher good outcome rates when combining elites
- higher resource gain from green and red crystals (+35%)
- higher resource gain from iron ore (+20%)
- higher resource gain from normal and red oil (+15%)
tech list (drone-specific):
- cloner: cloning is cheaper (-25%)
- scavenger, marauder: when delivered by them, scraps give more resources (+20%)
- crippler, marauder: +1 max targets
- destroyer: increased damage (+10%)
- prism: max-charged shot (4 prisms) have a high chance to inflict disarmed condition
- courier, trucker: extra payload capacity when carrying resources between the bases
- mortar: increased attack range (+15%)
