# Migration to the new package structure

The migration to the new package structure was done with the Bash script
`migrate.sh`. It can be used to repeat the migration after changes to the original structure.

Options:

- `--path` adapt file paths
- `--module` adapt module name (to ocm.software/ocm)
- `--migrator <scripfile>` create a migration script usable to migrate projects using this library

## Migrating projects using this library

The migration was done creating a migration script `migrate.mig`. It can be used
to migrate packages or whole projects using this library to the new package structure and module name.
It is executed by using the same migration Bash script `migrate.sh` using the option

`--script <migration script>` migrate a project using the OCM library.

Optionally, package paths can be added as additional arguments to the command line. Note that by default, 
if package paths are not specified, the complete current working directory is migrated, 
which might include `.git`, `.vscode` or any other subfolders you have there.

If you want to run the script on `macOS`, make sure you are using it with GNU `sed`.
