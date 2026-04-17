# AGENTS.md

## Skills
A skill is a set of local instructions to follow that is stored in a `SKILL.md` file.
Below is the list of repository-local skills that can be used in this project.

### Available skills
- `eco-knock-maintainer`: Maintain this repository while preserving its current architecture, naming, and layering. Use when adding features, refactoring packages, reorganizing directories, wiring gRPC or config code, or checking whether a change fits this repository's established style. (file: `./skills/eco-knock-maintainer/SKILL.md`)
- `git-commit-korean`: Inspect this repository's git history and current diff, then draft or create git commits that match the local convention. Use when the user asks to write a commit message, make a git commit, summarize changes into a commit, keep commit messages in Korean, or split changes into small logical commits aligned with recent repository history. (file: `./skills/git-commit-korean/SKILL.md`)
- `raspberry-pi-docker-deploy`: Deploy this repository to a Raspberry Pi with Docker, verify `.env.deploy` and `.env`, run the repository deploy scripts, and diagnose Raspberry Pi Docker or BME680/I2C issues such as missing `/dev/i2c-1` or failed `i2cdetect` results. Use when the user asks to deploy, redeploy, or debug Raspberry Pi execution. (file: `./skills/raspberry-pi-docker-deploy/SKILL.md`)
- `readme-maintainer`: Update this repository README so it stays aligned with the current implemented scope. Use when the user asks to create, rewrite, or refresh README content, document new features, update setup or deployment instructions, or keep README examples and limitations in sync with the current code and scripts. (file: `./skills/readme-maintainer/SKILL.md`)

## How to use skills
- Discovery: The list above is the repository-local skill registry for this project.
- Trigger rules: If the user names a skill directly, or the task clearly matches a listed skill, read that `SKILL.md` and follow it for the current turn.
- Trigger rules: If a listed skill seems even plausibly applicable, bias toward using it instead of hesitating or improvising an ad-hoc workflow.
- Communication: If you are using a skill for the current turn, explicitly say which skill you are using.
- Scope: Do not carry a skill across turns unless the user mentions it again or the next task still clearly matches it.
- Missing or blocked: If a listed skill file cannot be opened, say so briefly and continue with the best fallback.

## Local guidance
- Prefer repository-local skills in `./skills` before inventing ad-hoc workflow rules.
- If a task overlaps with a listed skill, do not skip it just because a non-skill approach also seems possible.
- When adding a new local skill under `./skills`, also add it to the `Available skills` list in this file so it can be auto-discovered in future sessions.
- Keep this file short. Put detailed task instructions in the skill's `SKILL.md`, not here.
