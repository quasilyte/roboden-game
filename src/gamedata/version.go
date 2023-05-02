package gamedata

// Season 0: build 11+
const SeasonNumber = 0

// Version 2
// - Added min movement radius to avoid accidental clicks
// - Fixed mismatching sprite size
//
// Version 3 (post jam)
// - Improved performance
// - Big map size option
// - New font
// - Updated drone cloning algorithm
// - Fixed a cloning price bug in colony planner
// - Fixed colony movement overjump
// - Fixed some typos
// - Fixed some text rendering issues (it was too blurry at times)
// - Made macos/osx build possible
// - Improved tutorial texts a bit
// - Improved colony planner & core code (agent search code)
// - Colonies now prefer to use green cloners
//
// # Version 4
//
// * Misc:
//   - Added Russian language support
//
// * New features:
//   - Added walls (in forms of land cracks)
//   - Added mountains as a higher height wall types
//   - Crawler creeps (spawned by a boss)
//   - Added mortar drone
//   - Added anti-air missile drone
//   - Added prism drone
//   - Added turrets
//
// * Visual improvements:
//   - Added a flash effect when a unit (friendly or not) takes damage
//
// * Gameplay:
//   - Unit cost changes (rebalancing)
//   - Creep bases are now guarded by a tower (usually)
//   - Higher resource priority decreases the upkeep (bonus caps at 80% priority with 60% reduction)
//   - Reworked most merge recipes
//   - Bases try to send at least 1 worker as reinforcements
//   - Base will try to produce at least 2 workers even if security priority is high
//   - Rework faction passive bonuses
//
// * Fixes:
//   - Drones no longer try to pick up a depleted resource
//   - Fixed an upkeed visual bug (trash is rendered near the base)
//   - Fixed invalid drone stats (max hp and speed)
//
// # Version 5
//
// * New features:
//   - Experimental mobile devices support
//   - Finished controllers (gamepad devices) support
//   - Crawlers now know a "scatter" behavior
//
// * UX:
//   - The colony selection is more precise now
//   - Added "toggle base" and "menu" (burger) buttons
//
// * Visual improvements:
//   - Added icons for the 5th option
//   - Reworked sprites for the action options
//
// * Fixes:
//   - Drones with charging mode no longer confuse creeps (we're clearing the waypoint now)
//   - Fixed crawlers "never scout" bug
//
// # Version 7 (was uploaded as version 6 by accident)
//
// * New features:
//   - Red crystals resource
//   - Drone ranks (normal, elite, super elite)
//
// * Gameplay
//   - Increased a colony drone limit
//   - Added more colony drone traits
//
// * UX:
//   - Made it clear which option was selected
//
// * Visual improvements:
//   - New action cooldown effect
//
// * Fixes:
//   - Upon defeat, hide menu and toggle buttons
//   - Fixed resource collection bug (drone cargo value stacking)
//
// * Performance:
//   - Use a pre-decoded ogg stream instead of decoding it on the fly
//   - Since 99% graphic objects are ge.Sprite, they're now stored as separate slices (less iface calls)
//
// # Version 8
// So many things happened, but I forgot to write them down.
//
// # Version 11
// * Added online leaderboard to the game
// * Added a separate rewards screen
// * Added teleports
// * Balance tweaks
// * Resources priority effect rework
// * Changed the default keyboard binds for camera to WASD
//
// # Version 12
//
// * New content:
//   - Added Scarab tier 2 drone
//   - Added Devourer tier 3 drone
//   - Added "Oil regeneration rate" world option
//
// * UX:
//   - Added difficulty description tag in the game lobby
//
// * Bug fixes:
//   - Fixed overlapping teleports bug
//
// * Optimizations:
//   - The game doesn't create animation objects for drones with no animation anymore
const BuildNumber int = 12
