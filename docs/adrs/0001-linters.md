# 1. Enabled linters

* Status: accepted
* Date: 2022-08-31
* Authors: @yitsushi
* Deciders: @yitsushi @mandelsoft

## Context

Our intention is to build and maintain a project where people can contribute
without learning a lot of unique coding rules. The code-base does not follow
common Go patterns and the project is huge already, so we can't just enable all
linters we want to use. We tried to find a point to start somewhere enabling
them.

## Decision

We decided to enable all linters with:

* no issues
* manageable amount of issues (no more than 20)
* auto-fixable issues (--fix flag)

All remaining linters is disabled until we decide to enable them.

## Consequences

We will file new issues about disabled linters and discuss if we want to enable
those linters.
