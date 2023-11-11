package gamedata

// Season 0: build 11+
// Season 1: build 14+
const SeasonNumber = 1

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
//   - Added "Terrain" world option
//   - Super creep versions (classic + arena)
//   - Added Seeker tier 2 drone
//
// * Balance:
//   - Add 2 max hp to Prism drones (28 -> 30)
//   - Make tether towers affect up to 4 drones (it was 2 before)
//   - Make tether towers slightly reduce the amount of drone energy consumption
//   - Made attack action range scale better
//   - Reworked (nerfed) couriers income
//
// * UX:
//   - Added difficulty description tag in the game lobby
//   - Added a UI toggle key
//   - Added a Timer extra option
//
// * Bug fixes:
//   - Fixed overlapping teleports bug
//   - Fixed universal drones not being queries by SearchWorkers in drones container
//   - Make "catch em all" achievement come with Elite grade right away
//   - Fixed multi rewards
//   - Fixed Trucker drone diode location
//   - Fixed leaderboard layout issues
//   - Removed "colony under attack" notice during the fatal damage
//   - Fixed some of the faction bonuses (they were not applied properly)
//
// * Optimizations:
//   - The game doesn't create animation objects for drones with no animation anymore
//   - Clusters for units (makes target selection faster)
//   - Less allocations in simulation mode
//   - Terrain is drawn as a single texture instead of a set of sprites
//
// * Visuals:
//   - Generate less looped land cracks.
//
// # Version 13 (pre-Steam itchio early access release)
//
// * New content:
//   - A completely new tutorial mode (the intro mission)
//   - Added Commander tier 2 drone
//   - Added Targeter tier 2 drone
//   - Added Firebug tier 2 drone
//   - Added Harvester turret
//
// * Difficulty settings:
//   - Up to 5 creep bases in Classic mode (the previous limit was 4)
//   - Enable arena progression setting in infinite arena mode
//   - Change arena-related settings scaling (20% -> 25%)
//
// * Balance:
//   - Increased the max upkeep resources cost
//   - Freighter: now has a zero upkeep
//   - Roomba: increase upkeep (3 -> 9)
//
// * Creeps:
//   - Added a 10-20 sec delay before a howitzer can start firing its artillery
//
// * UX:
//   - Fixed camera snapping issue near the borders of the map
//   - The message window is not half-transparent
//   - Better gamepad configuration screen
//   - Better drones overview (with stats)
//   - Added seed option help on-hover text
//   - A notification for the colony being destroyed
//   - Added PlayStation and Nintendo Switch gamepads layout support
//   - Auto-pause when gamepad is disconnected
//   - In-game on-hover hints
//   - Hide the less relevant cursor when in the game
//
// * Visual improvements:
//   - Prism drone attack now has a hit impact effect
//   - Better layers for colonies (especially when they're landing)
//   - Better layers for explosions (they're above dreadnought now, as they should be)
//   - Better layout in some of the menus
//
// * Gameplay:
//   - Reworked the tutorial
//
// * Bug fixes:
//   - "Colony" settings tab is not properly disabled in Reverse mode
//   - Splash screen no longer depends on the classic mode fog of war settings
//   - No achievements in coop mode (this would be too easy)
//   - Fixed uncentered camera in a single bot mode
//   - Fixed a Seeker drone that was unlocked right away (should be unlocked at 1000 pts)
//   - Can no longer build on top of the teleporter
//
// * Performance:
//   - Precompute waypoint direction once for colony agents
//
// * Misc:
//   - Updated ebitenui library version
//   - Underscore character '_' is now allowed in the username
//
// # Version 14 (Steam release)
//
// * New content:
//   - A new Secret achievement
//   - A new Terminal achievement
//   - A new Ark colony core
//   - A new music track (called Sexxxy Bit 3 by DROZERiX)
//   - A new Biome (Forest)
//   - A new tier 3 Bomber drone
//
// * Steam-related:
//   - Integrated Steam SDK, implemented Steam achievements
//   - "Steam" build label in the main menu
//   - Added some debug commands to the terminal
//
// * UX:
//   - The controller Home button now pauses the game too
//   - Added colony destination point marker
//   - Add an option to disable on-screen buttons
//
// * Visuals:
//   - Made VSync option configurable
//   - Brand new visual effects for many explosion-like animations
//   - Camera shaking (can be turned off)
//   - A textured beam for couriers
//
// * Misc:
//   - Added a logs.grep terminal command
//
// # Version 15
//
// * UX:
//   - Show Dreadnought health level in the tooltip
//   - Give an input device prompt for the new players
//   - Display drone recipe in the Drone Collection menu as well (even for T3 drones)
//   - Add drone highlighting when hovering over a recipe tab entry
//   - Auto-fill player's name from the Steam account (this can be changed in the menu later)
//
// * Controls:
//   - Reworked the default middle mouse button scroll behavior (can be reverted in the keyboard controls section)
//
// * New content:
//   - Add Spectator achievement
//   - Add 1337 achievement
//   - Add Gladiator achievement
//   - New Fortress creep (available via difficulty setting)
//   - New Templar creep
//   - New Ion Mortar creep turret (available via difficulty setting)
//   - A "gold" option that disables the gold resource deposits
//   - Added a Drone factory neutral building
//   - Added a Power plant neutral building
//   - Added a Turret neutral building
//   - Added "super creep rate" option in Reverse mode
//
// * Gameplay:
//   - Increase dreadnough max crawlers limit
//   - Added x2 game speed
//   - Made x1.2 game speed default (was x1.0, but it's kinda slow)
//   - Reworked colonies movement
//
// * Bug fixes:
//   - Fixed harvester en description
//   - Fixed dreadnought sprite flashing component
//   - Fixed devourer levels when joining a commander's group
//   - Fixed a intro mission crash when "keyboard" input method was selected
//   - Escape/back inside a results screen no longer skips through rewards
//   - Fixed a virtual cursor click over a screen buttons like "exit"
//   - Fixed the right/down edge scroll with its range set to 1
//
// * Visuals:
//   - Improved some effects layer arrangement
//   - Make screen shaking more intensive
//   - Added a damage shader to Dreadnought
//
// * Balance:
//   - Bombers now deal extra damage to buildings and dreadnought
//   - Increased Destroyer's evo points cost (11 -> 17)
//   - Increased Devourer's evo points cost (11 -> 12)
//   - Increased Guardians's evo points cost (8 -> 9)
//   - Make evo points generation slightly slower
//   - Some drones now use energy when attacking
//   - Unallocated drone point now grants 2 difficulty points instead of 1
//   - Dreadnought can now dispatch up to 7 crawlers (this limit was at 5 previously)
//   - Reduced Ark colony drone limit (100 -> 80)
//   - Wisps will retreat for longer distances
//   - Added damage vs buildings modifier
//   - Increased wave budget scaling in Infinite Arena mode
//   - Decrease "attack colony" action time cost in Reverse mode
//   - Scale the number of wisps with a map size
//
// * Computer player (colony bots):
//   - Will no longer play a solo base tactic with Ark core design
//   - Better max radius selection strategy
//   - Better evolution-related choice selection
//   - Uses Attack action more often with Firebug & Bomber drones
//
// * Misc:
//   - Trim trailing whitespace in the username
//   - New modes and options are now listed in the rewards screen
//
// * Tutorial:
//   - The difficulty & speed is now configurable via the terminal
//   - Pinned the special choice that is required to continue
//   - Removed the radius increase context hint (it's already in the main course)
//   - Removed the north attack notice (it was redundant)
//
// * Replays:
//   - Disabling replays from mismatching game versions
//   - Showing the game seed in the replay's description
//
// * Steam Deck:
//   - Showing the on-screen keyboard when appropriate
//   - Added Steam Deck layout option
//
// # Version 16 (Steam update 2)
//
// * New content:
//   - Added Inferno environment:
//     -- A new env-unique resource - Sulfur
//     -- A new env-unique resource - Volcanic rocks
//     -- A lava geyser trap
//     -- A lava lake that shoots magma projectiles
//   - A new tank-like ground colony
//   - A new siege turret
//
// * UX:
//   - Recipe tab tooltips now shows the T2 drone counters
//   - In the menu, pressing [B] on xbox and [O] on ps controllers now act as "back" buttons too
//
// * Balance:
//   - Difficulty % now affects Infinite Arena score more significantly
//   - Some resources are now less likely to be grouped together (like oil resource)
//   - Added some movement acceleration for the colonies; increased base max speed to compensate a bit
//   - Increased Power Plant resource yield (13 -> 15)
//
// * Visuals:
//   - Added some new tiles to the Moon environment
//   - Adjusted the Generator drone diode offset
//
// * Steam:
//   - Better Steam username auto-fill
//
// * Misc:
//   - Make Forest environment the default
//   - Howitzer creeps can no longer deploy near the map boundary
//   - Most ground units can now perform "diagonal moves"
//
// * Tutorial:
//   - Change the environment to Inferno
//   - Removed resources close to the base to improve the pacing
//
// * Computer player (colony bots):
//   - Less often moves away from the building that is being constructed
//   - Does less short moves when using a Den colony design
//   - Learned a few tricks about the Inferno environment (will avoid geysers when possible)
//   - Has increased interest in scraps if Scavengers/Marauders are available
//   - A better colony power calculation system
//
// * Optimization:
//   - Don't play a single sound more than once inside a single frame
//   - Changed the damage flashing; it was causing major performance issues on some machines
//   - Changed the color scaling method; now it's much more efficient
//
// * Bug fixes:
//   - Fixed a bug in long-range drones target seeking algorithm
//   - Don't show merge recipe for the locked drones
//   - Fixed the "creep spawn rate" difficulty option
//
// * Steam Deck:
//   - Set the default gamepad layout to Steam Deck when running a game there
//   - Like with other gamepads, [B] now goes back in the menus too
//
// # Version 17 (Steam update 3)
//
// * New features:
//   - Added in-game "fast forward" option
//   - Controllable Coordinators in Reverse mode
//   - Coordinators in Classic mode
//
// * New content:
//   - New Coordinator creep
//   - New Master Tactician achievement
//
// * UX:
//   - Show "game paused" message when the game is paused
//   - Show screen button hotkeys in their tooltip
//   - Tooltips are not disappearing as easily now
//   - Make Reverse mode tooltips more informative
//
// * Gameplay:
//   - Colonies without workers can now create a free worker in some cases
//
// * Balance:
//   - Re-balance some creeps and costs in Reverse mode
//   - Commander patrol radius is now lower, so its minions are not overshooting it
//   - Stealth crawlers: increased movement speed, increased burst damage at the cost of DPS
//   - Assault T3 creeps get a one-time damage shield
//   - Reworked multiple Reverse mode aspects
//
// * Visuals:
//   - Reworker Disintegrator projectile sprite a bit
//
// * Misc:
//   - Mark 450%+ difficulty as Ultimate Despair
//   - Reworked on-pause orders
//   - Stunner (ex. Templar) gets a unique super version bonus
//   - Level generator now saves the level checksum to make verification easier
//   - Keep only two levels of starting resources: none and full
//   - Changed lobby menu toggles design
//
// * Replays:
//   - Improved the replay save slot auto-selection
//
// * Bug fixes:
//   - Fix conflicting gamepad & mouse tooltips when playing in a split-screen mode
//   - Fix fog of war with Bastion bases
//   - Harvester no longer collects sulfur
//   - Fix Bastion colonies unstuck behavior
//   - Fix Harvester cell unmark on destruction
//
// * Computer player (colony bots):
//   - Added a new "comeback" move to bots
//   - Attack creep ground bases
//
// # Version 18 - couriers hotfix
//
// # Version 19
// - Storing more info inside a replay (should make the debugging easier)
// - Fix a bug where difficulty option is pressable even if it's locked
// - Updated Ebitengine version
// - Update ge package version
// - Add a pos correction for coordinator waypoints on the map boundary
//
// # Version 20
// - Exit prompt now does disable the fastforward too
// - Make oil regen rate worth of 10 difficulty points
// - Boost Discharger creep (speed up, hp up, attack range up)
// - Boost Coordinator creep (hp up)
// - Reduce the time needed to buy Stunners and Dischargers in reverse mode
// - Allow fast forward toggling via gamepad (LStick click)
// - Forbid fast forward in a multiplayer game
//
// # Version 21
// - New horizontal landscape mode
// - Try not to place a teleporter too close to the base starting pos
// - Improve creeps count scaling in the Reverse mode
// - Increase the max attack range when doing a Rally action in Reverse mode
// - Added "Into the Darkness" achievement
// - Added Megaroomba relict
// - Fixed a memory leak issue (solves a performance problem)
// - Added screen filters
//
// # Version 22
// - Added wasm loading spinner
// - Moved sfx back to embed data to make load times faster (even on desktops)
// - Added XM music support (optional for desktops, mandatory for wasm builds)
// - Fixed mobiles support on wasm builds
// - Include "clan" information to published results (Steam, itch.io, etc)
// - Anti Air bot point cost: 2 -> 3
// - Servo bot point cost: 3 -> 2
// - Fighter point cost: 3 -> 4
// - Destroyer gets DPS boost, but its attack energy cost is not higher as well
// - Increase Generator bot energy regen rate & mention it as special ability
//
// # Version 23
// - Den can now stomp a Howitzer as well
// - Improve computer bot danger score calculations
// - Forbid a very small map in Reverse mode
// - Tether Beam speed boost increased
// - Elite workers now have decreased upkeep costs
// - Increase most building costs (with an exception of Tether Beacon, it's cheaper now)
// - New Reverse mode difficulty toggle: Elite Fleet
// - Rebalance some of the difficulty calculations
// - Tier1 drones now have less maxEnergy
// - Nerf Prisms (~x2 energy cost per shot, makes Rechargers more synergetic)
// - Buff Den colony (building discount)
// - Buff Tether Beacon (increased range)
// - Buff Bomber (bombs deal more damage)
// - Buff some of the relicts
// - Added Grenadier creeps
// - Added Quicksilver achievement
// - Added Cheese achievement
// - Do not grant some achievements in the Inf Arena mode
// - Added "large diode" mode for accessibility
// - Update input library version
// - Show buildinfo tag on the main menu screen
// - Properly save empty platforms as "Steam"
// - Added wide screens support (18:9 through 21:9)
// - Added a 16:10 display ratio (useful for Steam Deck and tablets)
// - Added a on-screen keyboard support for Androids (because I can't make the native soft keyboard work everywhere)
// - Make "vertical" world shape fit the widest of resolutions to avoid camera issues
// - Added scaled (~x2) buttons support for mobile platforms
// - userdevice package now recognizes Android devices as mobile platforms
// - Using an updated game save/load package (gdata) to make it possible to save/load on Android
// - Embed XM music tracks (since they're available everywhere)
// - Force Android builds to use XM music player
// - Moved cmd/main implementation to cmd/internal/main to allow a second entry point (cmd/mobilegame)
// - Removed _localStorageAvailable from the index.html (it's not needed anymore)
// - Use new ge package version that uses mobile.SetGame instead of ebiten.RunGame for mobile platforms
// - Server can now achive interesting replays in addition to the rejected replays
const BuildNumber int = 23
