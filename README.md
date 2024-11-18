# settings - Imperative settings management for Go

## Key points

1. This module allows you to manage the application settings *imperative* way. Being not declarative is the
whole point of this project. Imperative code allows you to be more flexible and to have more control over the settings.
2. Performance is not a concern. The settings are usually loaded only once at the beginning of the application.
That's why we can afford to do a lot of unnecessary work in exchange for more flexibility.

If you don't share my unhappiness with the declarative style of managing settings, you probably don't need this module.

## Features

- Read configuration from YAML, env variables, and command line arguments, and combine them.
- Mark required settings.
- Set default values.
- Idempotently read configuration multiple times in different places (probably an anti-pattern, but sometimes useful).
