#!/usr/bin/env -S uv --quiet run --script
# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "click",
#     "textual",
#     "rich",
#     "PyYAML",
# ]
# ///

import os

import click

import lazyworktree.app as app_module
from lazyworktree.config import load_config
import lazyworktree.models as models
from lazyworktree.app import GitWtStatus


@click.command()
@click.option(
    "--worktree-dir",
    type=click.Path(file_okay=False, dir_okay=True),
    default=None,
    help="Override the default worktree root directory.",
)
@click.argument("initial_filter", nargs=-1)
def main(worktree_dir: str | None, initial_filter: tuple[str, ...]) -> None:
    config = load_config()
    resolved_dir = None
    if worktree_dir:
        resolved_dir = os.path.expanduser(worktree_dir)
    elif config.worktree_dir:
        resolved_dir = os.path.expanduser(config.worktree_dir)
    if resolved_dir:
        models.WORKTREE_DIR = resolved_dir
        app_module.WORKTREE_DIR = resolved_dir
    filter_value = " ".join(initial_filter).strip()
    app = GitWtStatus(initial_filter=filter_value, config=config)
    run_result = app.run()
    if run_result:
        click.echo(run_result)


if __name__ == "__main__":
    main()
