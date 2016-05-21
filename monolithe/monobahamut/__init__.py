# Author: Antoine Mercadal
# See LICENSE file for full LICENSE
# Copyright 2016 Aporeto.

from .writers.apiversionwriter import APIVersionWriter

__all__ = ['APIVersionWriter', 'plugin_info']


def plugin_info():
    """
    """
    return {
        'VanillaWriter': None,
        'APIVersionWriter': APIVersionWriter,
        'PackageWriter': None,
        'CLIWriter': None,
        'get_idiomatic_name': None,
        'get_type_name': None
    }
