#!/bin/bash
TESTS_RUN=0
TESTS_SUCCESS=0
for example in examples/*.hlisp;
do
  TESTS_RUN=$((TESTS_RUN+1));
  echo "${example}" && ./heisenlisp --script "${example}" && echo OK "${example}" && TESTS_SUCCESS=$((TESTS_SUCCESS+1));
  echo;
done

echo "${TESTS_SUCCESS}/${TESTS_RUN} PASS"
if [ "${TESTS_SUCCESS}" -ne "${TESTS_RUN}" ]; then
  echo "FAILURE"
else
  echo "SUCCESS"
fi
