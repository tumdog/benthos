Grok
====

This package is a fork of github.com/trivago/grok, which is itself a fork of
github.com/vjeantet/grok.

The intention of this fork is to allow build tags to replace the default Go
regex library with alternative C library regex implementations.

The API has also been largely cut down, and now only supports the ParseTyped
methods
