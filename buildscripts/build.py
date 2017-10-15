#!/usr/bin/env python3

import subprocess
import logging

FORMAT = '%(asctime)-15s %(message)s'
logging.basicConfig(format=FORMAT, level=logging.INFO)

def call(*args):
    logging.info("Running: {}".format(args))
    p = subprocess.Popen(args, stdout=subprocess.PIPE)
    stdout, _ = p.communicate()
    logging.info("Done: {}".format(args))
    return stdout.decode("utf-8").strip()

def spaces_to_triple_underscores(x):
    return x.replace(" ", "___")

def get_version_string():
    return call("git", "describe", "--long", "--dirty", "--abbrev=10", "--tags")

def get_commit_hash():
    return call("git", "rev-parse", "HEAD")

def get_go_version():
    return call("go", "version")

def get_datetime_string():
    return call("date", "--iso-8601=s")

def get_build_machine_info():
    return call("uname", "-mnopv")

def build_versiondict():
    return {
        "VersionString": get_version_string(),
        "CommitHash": get_commit_hash(),
        "BuildTimestampISO8601": get_datetime_string(),
        "GoVersion": get_go_version(),
        "BuildMachineInfo": get_build_machine_info(),
    }

def build(versiondict):
    ldargs = []
    for k, v in versiondict.items():
        k = "github.com/steinarvk/heisenlisp/version." + k
        v = spaces_to_triple_underscores(v)
        ldargs.append("-X " + k + "=" + v)
    args = []
    args.append("--ldflags")
    args.append(" ".join(x for x in ldargs))
    call("go", "build", *args)

def main():
    build(build_versiondict())

if __name__ == "__main__":
    main()
