############
 DBGP Tools
############

This repository contains a set of tools and libraries to interact with
`DBGP <https://xdebug.org/docs/dbgp>`_, the_protocol that `Xdebug
<https://xdebug.org>`_ uses for communication with IDEs.

There are three tools:

************
 dbgpClient
************

A command line debug client allows you to debug DBGP-supported languages
without having to set up an IDE.

The usage documentation is at https://xdebug.org/docs/dbgpClient

***********
 dbgpProxy
***********

This tool allows you to proxy and route debugging requests to IDEs
depending on which IDE key is in use.

The usage documentation is at https://xdebug.org/docs/dbgpProxy

***********
 xdebugctl
***********

A tool that allows you to control Xdebug out-of-band of a PHP request.

This usage documentation is at https://xdebug.org/docs/xdebugctl
