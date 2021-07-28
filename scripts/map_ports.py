#!/usr/bin/env python3
import argparse
import re
import os
import subprocess
import sys

STACK_DIR = os.path.expanduser(os.path.join('~', '.firefly', 'stacks'))
HOST_PORT_PATTERN = re.compile(r'([A-Za-z0-9._-]+):(\d{4})')

def _run_docker_cmd(args):
  p = subprocess.run(['docker'] + args, capture_output=True)
  return p.stdout.decode().rstrip()

def _find_docker_container(name):
  return _run_docker_cmd(['ps', '--filter', 'name=' + name, '--format', '{{.ID}}'])

def _find_docker_port(container_id, container_port):
  output = _run_docker_cmd(['port', container_id])
  for line in output.splitlines():
    source, dest = line.split(' -> ')
    source_port = re.search(r'(\d+)/', source).group(1)
    dest_port = re.search(r':(\d+)', dest).group(1)
    if source_port == container_port:
      return dest_port
  return None

def main():
  parser = argparse.ArgumentParser(
    description=(
      'Read a FireFly config file and map internal Docker ports to the '
      'exposed ports on localhost.'
    ))
  parser.add_argument('name', help='name of a running FireFly CLI stack')
  parser.add_argument(
    '-i', '--index', help='0-based index of container in the stack (default=0)',
    type=int, default=0, required=False)
  args = parser.parse_args()

  filename = os.path.join(
    STACK_DIR, args.name, 'configs', 'firefly_core_%d.yml' % args.index)

  with open(filename) as f:
    for line in f.readlines():
      match = HOST_PORT_PATTERN.search(line)
      if match:
        host = match.group(1)
        port = match.group(2)
        if host != 'localhost' and host != '127.0.0.1':
          container_id = _find_docker_container(host)
          if container_id:
            mapped_port = _find_docker_port(container_id, port)
            if mapped_port:
              line = line.replace(
                '%s:%s' % (host, port), 'localhost:%s' % mapped_port)
      print(line, end='')

if __name__ == '__main__':
  main()
