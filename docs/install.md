# Setup thingsleilasays

## Create Heroku app

Create a heroku app:

```bash
heroku create thingsleilasays
```

## Create S3 bucket

Setup the AWS CLI:

```bash
aws configure
```

### Create a new S3 bucket for tweets

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
                "NoncurrentDays": 3
            },
            "ID": "Tweet bucket lifecycle configuration"
        }
    ]
}'
```

### Create an IAM user and IAM profile for access to S3

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

### Create credentials and setup thingsleilasays to use them

Create an access key for the `thingsleilasays` user:

```bash
aws iam create-access-key --user-name thingsleilasays
```

Copy the access key ID and the secret access key and set them as config vars
on the `thingsleilasays` app:

```bash
heroku config:set -a thingsleilasays \
    AWS_REGION=us-east-1 \
    AWS_ACCESS_KEY_ID=$ACCESS_KEY_ID \
    AWS_SECRET_ACCESS_KEY=$SECRET_ACCESS_KEY \
    S3_BUCKET=thingsleilasays
```

At this point the `thingsleilasays` app should have access to the read and
write tweets to the S3 bucket.

## Create Twitter app

Login to Twitter, go to https://apps.twitter.com and click *Create New App*.

### Create app

Fill the form, check the developer agreement box and click *Create your
Twitter application*.

### Setup permissions

Visit the app and click on the *Permissions* tab. Select *Read only* and click
*Update Settings*.

### Fetch access tokens

CLick on the *Keys and Access Tokens* tab. Copy the consumer key, consumer
secret, access token and access token secret and set them as config vars on
the `thingsleilasays` app:

```bash
heroku config:set -a thingsleilasays \
    TWITTER_CONSUMER_KEY=$CONSUMER_KEY \
    TWITTER_CONSUMER_SECRET=$CONSUMER_SECRET \
    TWITTER_ACCESS_TOKEN=$ACCESS_TOKEN \
    TWITTER_ACCESS_SECRET=$ACCESS_TOKEN_SECRET \
    TWITTER_USERNAME=$USERNAME
```

## Deploy Heroku app

Push the code to Heroku and scale a web dyno

```bash
git push heroku master
heroku scale web=1:Free
```

## Setup Heroku Scheduler

Add the scheduler addon:

```bash
heroku addons:create scheduler:standard
```

Open it:

```bash
heroku addons:open scheduler
```

Click *Add new job*, enter `fetch` in the command field, set the job frequency
to *Every 10 minutes* and click *Save*. Run the command once to seed the
dataset:

```bash
heroku run fetch
```
