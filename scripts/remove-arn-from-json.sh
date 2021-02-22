#!/bin/bash

jq 'del(.manifest.apis.custom.endpoint)' skill-package/skill.json > skill-package/skill.json.tmp

mv skill-package/skill.json.tmp skill-package/skill.json
