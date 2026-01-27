#!/bin/bash
# Quick SonarQube check script
# Usage: ./scripts/sonarqube-check.sh [path]

TARGET_PATH=${1:-.}
PROJECT_KEY="ai-pack"

echo "Running SonarQube analysis on: $TARGET_PATH"
cd "$TARGET_PATH" || exit 1

# Run sonar-scanner with parent config
sonar-scanner \
  -Dsonar.projectBaseDir="$(pwd)" \
  -Dproject.settings="../sonar-project.properties" 2>&1 | tee /tmp/sonar-output.txt

echo ""
echo "Analysis complete. Check SonarQube Cloud for results."
echo "https://sonarcloud.io/project/overview?id=$PROJECT_KEY"
