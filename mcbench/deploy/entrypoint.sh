#!/bin/sh
# Sets up git for the worker, then runs the command. It authenticates with the
# GitHub deploy key in GITHUB_DEPLOY_KEY, checks GitHub against the host keys
# baked into the image, and clones the repository once onto the volume, at
# online.repo (MCBENCH_ONLINE_REPO).
set -eu
key="$HOME/.ssh/deploy_key"
mkdir -p "$HOME/.ssh"
(umask 077 && printf '%s\n' "${GITHUB_DEPLOY_KEY:?set GITHUB_DEPLOY_KEY to a deploy key with write access}" > "$key")
unset GITHUB_DEPLOY_KEY
export GIT_SSH_COMMAND="ssh -i $key -o IdentitiesOnly=yes -o StrictHostKeyChecking=yes -o UserKnownHostsFile=/etc/ssh/github_known_hosts"

repo="${MCBENCH_ONLINE_REPO:?set MCBENCH_ONLINE_REPO to the clone on the volume}"
if [ ! -d "$repo/.git" ]; then
	git clone --quiet --no-checkout "${GITHUB_REPO_URL:-git@github.com:KanarekLife/ResearchProject.git}" "$repo"
fi
git config --global user.name "${GIT_USER_NAME:-mcbench-worker}"
git config --global user.email "${GIT_USER_EMAIL:-mcbench-worker@users.noreply.github.com}"
exec "$@"
