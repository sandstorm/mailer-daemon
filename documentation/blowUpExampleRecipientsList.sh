#!/bin/bash

cat exampleRecipientsList.json > largeExampleRecipientsList.json
for i in {1..100000}
do
   cat exampleRecipientsList.json | sed "s/@/$i@/g" >> largeExampleRecipientsList.json
done