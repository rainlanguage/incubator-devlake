const github_project = [
  [
    {
      plugin: 'github_graphql',
      subtasks: [
        "Collect Projects",
        "Extract Projects"
      ],
      options: {
        connectionId: 1,
        owner: "rainlanguage",
        projectNumber: 4
      },
    },
  ],
];


export default github_project;
