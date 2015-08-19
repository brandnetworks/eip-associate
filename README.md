# EIP Associate

Associates an elastic ip with an ec2 instance, picking a free one from a predefined list

## Running

    docker brandnetworks/eip-associate --eips eips

## Usage

    Usage: eip-associate --eips eips
      -eips="": Comma separated list of elastic ips
      -metadata="http://169.254.169.254/latest/meta-data": Meta data endpoint
      -pause=5: Number of seconds to pause between retries
      -retries=10: Maximum number of retries

## Building

    docker build -t builder . && docker run builder | docker build -t eip-associate -
