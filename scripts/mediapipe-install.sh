#!/bin/bash
cd ..
python3 -m venv ./.venv # REQUIREMENTS: python3, pip3 must be installed
source ./.venv/bin/activate
cd python
pip3 install -r requirements.txt