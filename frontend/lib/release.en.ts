import type { UpdateNote } from './release';

// Notas de versão em inglês, na mesma ordem de UPDATE_NOTES (um teste confere a
// versão de cada posição). Ao publicar uma versão nova, escreva as duas.
export const UPDATE_NOTES_EN: readonly UpdateNote[] = [
  {
    version: '0.38.0',
    title: 'Cadência now speaks English too',
    description:
      'At the top of every screen, and in Settings, you can choose between Portuguese and English. Your choice applies to the screens, the workouts and their explanations, the notices and the emails. The Terms of Use and the Privacy Policy also got a translation; the Portuguese version is still the one that legally applies. The language is remembered on the device where you chose it.',
  },
  {
    version: '0.37.0',
    title: 'Workouts that follow your recovery',
    description:
      'When you report pain or high fatigue, the plan now protects itself in levels (light, moderate or strong) instead of freezing the whole cycle. The protection has an end date, eases as the days go by and is reassessed after every completed workout and every check-in. On the plan screen, the notice explains the reason and how long it lasts, and the "I\'ve recovered" button asks for a new assessment right away. If you had pain in the last few days, the protection stays on for safety.',
  },
  {
    version: '0.36.0',
    title: 'Link an imported activity to a workout',
    description:
      'In Activities > Import, each activity can now be linked to the workout planned for the same day (or a nearby day), and the link shows in the list with the workout name. The link only records the relationship: it does not change the plan or the workout. You can also link activities you had already imported and undo the link whenever you want.',
  },
  {
    version: '0.35.0',
    title: 'A new look, inspired by trail maps',
    description:
      "Cadência got its own identity, inspired by MTB and gravel topographic maps. Today's workout is drawn as a route, and intensity uses trail signage: green circle for easy, blue square for moderate and black diamond for hard. Navigation is now the same on every screen and, on phones, sits in a bottom bar within thumb's reach.",
  },
  {
    version: '0.34.0',
    title: 'Import your real activities',
    description:
      'In Activities, you can now import the .fit or .gpx file from your watch or bike computer (Garmin, Wahoo, XOSS and others). Cadência shows the summary — duration, distance, elevation, heart rate, power and cadence — and suggests the workout planned for the same day; you confirm whether to use that data when you complete the session. On Android, install Cadência on your home screen to share the file straight from your device app.',
  },
  {
    version: '0.33.0',
    title: 'Safer account access',
    description:
      'In Settings you can now change your password and sign out of other devices. When you change the password, other devices are signed out automatically. We also made Cadência more stable: slow network requests are retried safely and, if your session expires, you are taken back to sign-in.',
  },
  {
    version: '0.32.0',
    title: 'A fuller and safer cycling onboarding',
    description:
      'Your profile can now record more cycling context, availability, history, equipment, FTP and safety signs. The adaptive questionnaire asks each question at the right moment, while optional measurements and progress indicators stay observational and do not change the plan without enough evidence.',
  },
  {
    version: '0.31.0',
    title: 'Settings and safe account deletion',
    description:
      'The new settings area brings together your profile and account data. If you decide to delete your account, Cadência asks for your password and an explicit confirmation before permanently erasing your data from the database.',
  },
  {
    version: '0.30.0',
    title: 'Post-event protection and sounder adaptation',
    description:
      'After an event you entered, the plan can prioritize up to seven days of easy recovery, with conservative duration and RPE. The ride log also accepts average cadence as optional observational data. Each session now also shows the data, restrictions and alternatives the rule considered. Incomplete feedback, invalid data and recent protective signs no longer change the load automatically; limitations and returning after a break still take priority.',
  },
  {
    version: '0.29.0',
    title: 'Safety and equipment context',
    description:
      'Post-workout feedback can now record the equipment you used. Your profile also lets you report warning symptoms and medical restrictions; this data reinforces the safety blocks and does not replace a professional assessment.',
  },
  {
    version: '0.28.0',
    title: 'Controlled threshold pilot',
    description:
      'Eligible advanced athletes can choose Threshold as a preference for a conservative three-block session. Cadência requires a road or indoor context, a passed assessment and a minimum history; the session uses RPE, sets no universal power target and keeps every pain and recovery protection.',
  },
  {
    version: '0.27.0',
    title: 'Clearer training status',
    description:
      'In your profile, you can say whether you are training regularly or returning after a break. Only the return option triggers easy continuous sessions of up to 45 minutes at RPE 3.5; the number of weeks alone does not reduce the plan.',
  },
  {
    version: '0.26.0',
    title: 'Catalog with an explicit long ride',
    description:
      "The cycle's longest session now shows as Long ride, with an endurance structure and a conservative load. The name sets the goal apart without automatically increasing duration, RPE or frequency.",
  },
  {
    version: '0.25.0',
    title: 'Fuller post-workout feedback',
    description:
      'When you complete a ride, you can record satisfaction, terrain and outside conditions in structured fields. This data appears in your history and remains observational, without changing the plan automatically.',
  },
  {
    version: '0.24.0',
    title: 'Safe correction of ride metrics',
    description:
      'Post-workout logs flagged for review can now have their optional metrics corrected without changing duration, RPE, feedback or the plan. The original value is kept for auditing.',
  },
  {
    version: '0.23.0',
    title: 'Clearer inconsistent logs',
    description:
      "When the ride's time and metrics don't match, the log is still kept, but the app warns that it will not be used when reading your history.",
  },
  {
    version: '0.22.0',
    title: 'Clearer access errors',
    description:
      'Temporary API or database problems now show as a loading error on signed-in screens. Cadência only returns to sign-in when the session really is not authenticated.',
  },
  {
    version: '0.21.0',
    title: 'Fuller safety context',
    description:
      'When you report a limitation, you can record its location, perceived intensity, what makes it worse and when it started. This data helps keep a safer reading and does not replace a professional assessment.',
  },
  {
    version: '0.20.0',
    title: 'Profile always reachable from the menu',
    description:
      'The side navigation now fits small screens better and keeps access to your profile out of the area covered by the footer.',
  },
  {
    version: '0.19.0',
    title: 'Better organized release history',
    description:
      "The initial notice got shorter and you can now browse every update on the new What's new page, also available in the menu.",
  },
  {
    version: '0.18.0',
    title: 'Post-workout feedback with more context',
    description:
      'When you complete a workout, you can record how recovered you feel and your confidence to repeat the session. These signs are observational and do not yet change the load automatically.',
  },
  {
    version: '0.17.0',
    title: 'Logging a partial completion',
    description:
      'When you complete a workout, you can say whether you did the whole session or only part of it and give the reason. This context helps the app read what you did without increasing the load automatically.',
  },
  {
    version: '0.16.0',
    title: 'Clearer discipline scope',
    description:
      'Cadência stays focused on road, MTB XCO, gravel, XCM and indoor. Downhill/enduro and track sprint/BMX are not part of this app and are not accepted as training disciplines.',
  },
  {
    version: '0.15.0',
    title: 'Short self-regulated intervals pilot',
    description:
      'Eligible advanced athletes who choose short intervals can get, on the road or indoors, six controlled 1-minute repetitions with easy recovery. The app keeps RPE, the minimum history and the pain and recovery protections, without allowing all-out sprints.',
  },
  {
    version: '0.14.0',
    title: 'VO₂max intervals pilot for road',
    description:
      'Eligible advanced athletes who choose VO₂max can get four controlled 4-minute blocks, based on recent evidence. The app keeps RPE, the minimum duration and the pain, recovery and history protections.',
  },
  {
    version: '0.13.0',
    title: 'Event-driven pre-race taper',
    description:
      'Eligible advanced athletes with an upcoming event can get a conservative reduction in volume, keeping the planned frequency and the pain and recovery protections.',
  },
  {
    version: '0.12.0',
    title: 'Planning by how close the event is',
    description:
      'When a future event is entered, the specific phase is only used in the period close to the race; distant events keep the regular progression.',
  },
  {
    version: '0.12.0',
    title: 'Adapted sessions can be started more safely',
    description:
      'A session adjusted for recovery can be started normally, while the pain, fatigue and recovery protections stay in place.',
  },
  {
    version: '0.12.0',
    title: 'Access and feedback protections',
    description:
      "The app strengthens the protection of sign-in routes, avoids concurrent processing of the weekly summary and only marks what's new as seen after you dismiss it.",
  },
  {
    version: '0.11.1',
    title: 'Recovery week without a quality session',
    description:
      "The cycle's fourth week keeps the long ride and uses active recovery in the other sessions, with no race pace or quality intervals.",
  },
  {
    version: '0.11.0',
    title: 'High-intensity intervals pilot for road',
    description:
      'Eligible advanced athletes can get, in alternating cycles, a high-intensity interval session based on recent evidence. The app keeps conservative limits and the pain and recovery protections.',
  },
  {
    version: '0.10.0',
    title: 'Logging a missed workout',
    description:
      'Past workouts that were not done can be logged explicitly. The app does not create an automatic make-up session or increase the next load because of that log.',
  },
  {
    version: '0.9.0',
    title: 'Active recovery in the recovery week',
    description:
      'The plan can alternate an active recovery session, with easy effort and reduced volume, without replacing the protections applied when there is pain, a limitation or insufficient recovery.',
  },
  {
    version: '0.8.0',
    title: 'Aerobic intervals pilot for MTB XCO',
    description:
      'Advanced athletes with a passed submaximal assessment, a compatible goal and an explicit XCO context can get an aerobic pilot with controlled blocks, without all-out sprints or technical race simulation.',
  },
  {
    version: '0.7.0',
    title: 'Evidence-based cycling catalog',
    description:
      'Scientific references are now organized by discipline and support eligible protocols, starting with the moderate road intervals pilot.',
  },
  {
    version: '0.6.0',
    title: 'Observed context in the plan',
    description: 'When you generate a new plan, you can see which recent records helped set a more suitable progression.',
  },
  {
    version: '0.6.0',
    title: 'Feedback and recovery with more context',
    description:
      'Your completed workouts, perceived effort and recovery check-ins help keep your next sessions safer.',
  },
  {
    version: '0.6.0',
    title: 'Ride log',
    description: 'You can record distance, elevation, heart rate and power with the workout, when you have that data.',
  },
  {
    version: '0.6.0',
    title: 'Explainable workouts',
    description:
      'Each session keeps showing its structure, the reason it was chosen and the important precautions for doing the workout.',
  },
] as const;
