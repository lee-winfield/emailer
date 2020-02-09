#!/bin/bash

BUCKET="cjwinfield" # bucket name
FILENAME="email.zip" # upload key
OUTPUT_FOLDER="./build" # will be cleaned

HERE=${BASH_SOURCE%/*} # relative path to this file's folder
OUTPUT_FILE="$OUTPUT_FOLDER/$FILENAME"

# create target folders
mkdir $OUTPUT_FOLDER

# Create Binary
GOOS=linux GOARCH=amd64 go build -o email/email ./email
chmod 755 ./email

# zip everything to output folder (recursively and quietly)
echo "zipping project"
zip -j $OUTPUT_FILE ./email/email


# # upload to S3
echo "Uploading to S3"
aws s3 cp --acl public-read $OUTPUT_FILE s3://$BUCKET/$FILENAME
echo "https://s3.amazonaws.com/$BUCKET/$FILENAME"

# clean everything
echo "Cleaning"
rm -rf $OUTPUT_FOLDER

# aws lambda update-function-code

aws lambda update-function-code --function-name EmailFunction --s3-bucket "${BUCKET}" --s3-key "${FILENAME}"
echo "Done"