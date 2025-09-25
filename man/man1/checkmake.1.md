---
title: checkmake(1) checkmake User Manuals | checkmake User Manuals
author: Daniel Schauenberg <d@unwiredcouch.com>
date: REPLACE_DATE
---

# NAME
**checkmake** -- linter for Makefiles

# SYNOPSIS

**checkmake** \[options\] makefile ...

# DESCRIPTION
`checkmake` is a linter for Makefiles. It allows for a set of
configurable rules being run against a Makefile or a set of `\*.mk` files.

# OPTIONS

**-h**, **--help**
:    Show a friendly help message.

**--version**
:    Show version.

**--debug**
:    Enable debug mode

**--config=\<configPath\>**
:    Configuration file to read

**--format=\<format\>**
:    Output format as a Golang text/template template

**--list-rules**
:    List registered rules

# CONFIGURATION
By default checkmake looks for a `checkmake.ini` file in the same
folder it's executed in, and then as fallback in `~/checkmake.ini`.
This can be overridden by passing the `--config=` argument pointing it
to a different configuration file. With the configuration file the
`[default]` section is for checkmake itself while sections named after
the rule names are passed to the rules as their configuration. All
keys/values are hereby treated as strings and passed to the rule in a
string/string map.

The following configuration options for checkmake itself are supported within
the `default` section:

**default.format**
:    This enables the custom output formatter with the given template string
as a format

maxBodylength.maxBodylength
    This allows to override the maximum number of lines for a rule body
    that checkmake will allow from the default of 5  to a different number

minphony.required
    This allows to override the list of minimum required phony targets
    from the default of (all, test, clean) to any list of target name strings.
    The value is a comma-separated list of strings.
    Setting minphony.required to the empty string disabled the minphony rule altogether.



# EXIT STATUS
checkmake exits with the following status:

```
 0:   checkmake ran successfully and found 0 violations
>1:   checkmake found the number of violations reflected by the exit status
```

In addition to checkmake having found 1 violation, exit status 1 is also used
to denote an error in execution happening.

# BUGS
Please file bugs against the issue tracker:
https://github.com/checkmake/checkmake/issues

# SEE ALSO
make(1)
