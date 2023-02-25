package menus

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
//
// * Visual improvements:
//   - Added a flash effect when a unit (friendly or not) takes damage
//
// * Gameplay:
//   - Unit cost changes (rebalancing)
//   - Creep bases are now guarded by a tower (usually)
//   - Higher resource priority decreases the upkeep (bonus caps at 80% priority with 60% reduction)
//   - Reworked most merge recipes
//
// * Fixes:
//   - Drones no longer try to pick up a depleted resource
//   - Fixed an upkeed visual bug (trash is rendered near the base)
const buildNumber int = 4
