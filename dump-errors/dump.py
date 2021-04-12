import urllib.request
import sys
import imp
import re
import inspect
import json

ERRORS_PY_URL = "https://raw.githubusercontent.com/freeipa/freeipa/34600a0ecac3ad3fbe7b7b5767c3a4c1a455dc45/ipalib/errors.py"

import_regex = re.compile(r"^(from [\w\.]+ )?import \w+( as \w+)?$")


def should_keep(l):
    return (import_regex.match(l) is None)


errors_py_str = urllib.request.urlopen(ERRORS_PY_URL).read().decode('utf-8')
errors_py_str = "\n".join(
    [l for l in errors_py_str.splitlines() if should_keep(l)])
errors_py_str = """
class Six:
    PY3 = True
six = Six()
ungettext = None
class Messages:
    def iter_messages(*args):
        return []
messages = Messages()
""" + errors_py_str

errors_mod = imp.new_module('errors')
exec(errors_py_str, errors_mod.__dict__)

error_codes = [
    {
        "name": k,
        "errno": v.errno
    } for k, v in inspect.getmembers(errors_mod)
    if hasattr(v, '__dict__') and type(v.__dict__.get('errno', None)) == int
]
error_codes.sort(key=lambda x: x["errno"])

with open('../data/errors.json', 'w') as f:
    json.dump(error_codes, f)
