# Create S3 bucket

Setup the AWS CLI:

```bash
aws configure
```

## Create a new S3 bucket for tweets

Create a new bucket in us-east-1:

```
aws s3api create-bucket --bucket thingsleilasays --region us-east-1
```

Enable bucket versioning:

```
aws s3api put-bucket-versioning --bucket thingsleilasays --versioning-configuration Status=Enabled
```

Setup bucket lifecycle configuration:

```
aws s3api put-bucket-lifecycle-configuration --bucket thingsleilasays --lifecycle-configuration '{
    "Rules": [
        {
            "Status": "Enabled",
            "Prefix": "",
            "NoncurrentVersionExpiration": {
                "NoncurrentDays": 14
            },
            "ID": "Domain snapshot bucket lifecycle configuration"
        }
    ]
}'
```

## Create an IAM user and IAM profile for access to S3

Create a new IAM user for the thingsleilasays app:

```bash
aws iam create-user --user-name thingsleilasays
```

List the users to confirm that the `thingsleilasays` user was created:

```bash
aws iam list-users
```

Create and attach an IAM policy to the `thingsleilasays` IAM user to grant
read/write access to the S3 bucket:

```
aws iam put-user-policy --user-name thingsleilasays --policy-name thingsleilasays-tweets --policy-document '{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject"
            ],
            "Resource": [
                "arn:aws:s3:::thingsleilasays/tweets.json"
            ]
        }
    ]
}'
```

List the user policies to confirm that `thingsleilasays-tweets` is attached
to the `thingsleilasays` IAM user:

```bash
aws iam list-user-policies --user-name thingsleilasays
```

## Create credentials and setup thingsleilasays to use them

Create an access key for the `thingsleilasays` user:

```bash
aws iam create-access-key --user-name thingsleilasays
```

Copy the access key ID and the secret access key and set them as config vars
on the `thingsleilasays` service:

```bash
heroku config:set -a thingsleilasays \
    AWS_REGION=us-east-1 \
    AWS_ACCESS_KEY_ID=$ACCESS_KEY_ID \
    AWS_SECRET_ACCESS_KEY=$SECRET_ACCESS_KEY
```

At this point the `thingsleilasays` app should have access to the read and
write tweets to the S3 bucket.
