read the specification on docs/backend and please complete the tasks until its done. for each completion of a feature, please push to the github repo.

You are an AI software agent working inside an existing Git repository.

Objective

Understand the project and write user documentation with screenshots if needed using chrome devtools mcp.

Workflow Rules (Very Important)

Work incrementally, one small step at a time

Do not try to understand everything at once

After each meaningful checkpoint, do:

Update documentation (AGENTS.md and/or TODO.md)

Commit changes

Push to the remote repository

Keep commits small, descriptive, and frequent

Never skip documentation updates

Avoid loading or reasoning about files not needed for the current step


Task Discovery (Incremental)



Inspect the /docs directory



Read one document at a time



For each document:



Extract tasks, requirements, or implementation steps



Do not attempt implementation yet



Append discovered tasks to TODO.md using this format:



## <Doc Name>

- [ ] Task description





After each document:

make sure you have appropriate .gitignore



Commit and push



Commit message example:



docs: extract tasks from <doc-name>



Step 3 — Task Breakdown



Review TODO.md



Break large tasks into small, actionable steps

Make sure it runs on my local machine first, connect to local mongodb etc

acceptance criteria are:
- all functions should work correctly
- should create unit tests and all passed
