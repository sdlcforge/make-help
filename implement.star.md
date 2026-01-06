## Setup isolated working direcotries for each task

For each task you implement, create a new branch and worktree.
1. Determine the branch name:
   - If the work is based on a phase+task plan, the branch and worktree dir should  be named 'phase-<phase #>-task-<task #>-<short change summary>', e.g.: 'phase-02-task-03-log-login-attempts'
   - Otherwise, use the date plus summary: '<YYYY-MM-DD>-<short change summary>', e.g.: '2025-01-15-log-login-attempts'
2. `mkdir -p worktrees` if necessary; *important*: worktrees are created in the `worktrees` directory.
3. Create the worktree: `git worktree add ./worktrees/<branch name> -b <branch name>`, e.g.: `git worktree add ./worktrees/2025-01-15-log-login-attempts -b 2025-01-15-log-login-attempts`.
4. If there is a `package.json` in the project root, run `npm -i` in each worktree directory.

## Use agents to implement changes

- Implement tasks in parallel where possible.
- Detremine the appropriate development agent based on the source (node-dev for Javascript and Typescript or golang-dev for go code).
- Provide the development agent with references to relavent files (always relative to the worktree directory) and detailed instructions on the task (if the instructions are already in a file, reference the file plus any section and/or line numbers; do not repeat existing instructions).
- Instruct the development agent to stop and ask questions if anything is unclear or the agent gets stuck or if things just aren't working (~5+ attempts fail unexpectedly).
- Set the context/working dir for agents assigned to this task to the task worktree directory.

## Review and followup on changes

- After the development agent finishes initial implementation, give the code-reviewer agent the same file references and task description and ask it to review the changes.
- Give the code-reviewer the git command to see the changes (`git diff <root work branch>`),
- Set the code-reviever agent's working directory to the task worktree directory.
- If the code-reviewer suggests any changes, launch a new development agent to consider all the suggested changes and implement them unless it disagrees with or does not fully understand the suggestion.
- Any suggested changes not implemented must be reported to the user.

## Integrating changes

- Once all suggestions (if any) are implemented, if the user requested to review changes before committing or there is some question or uncertaintity in the correctness of the implementation then stop and tell the user. Otherwise, review is not necessary and the changes may be merged.
- When ready to merge, merge the changes to the main workbranch (the branch the root project is on) or other worktree task branches (fixes that effect multiple in-flight tasks). Use the `--no-ff` in the merge command.
- If there are any conflicts, step and let the user resolve them.
- Once the merge is complete, remove the worktree (`git worktree remove <path to worktree>`) and delete the branch (`git branch -d <branch name>`).

## Update plans

- If the task was documented in a file (as a standalone task description, part of a TODO list, etc.) update the file to indicate the task is complete in the main workbranch.
