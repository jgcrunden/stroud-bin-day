![Github Actions Status Badge](https://github.com/jgcrunden/stroud-bin-day/actions/workflows/go.yml/badge.svg)
![Github Actions Status Badge](https://github.com/jgcrunden/stroud-bin-day/actions/workflows/ask-cli.yml/badge.svg)

# stroud-bin-day
An Alexa skill which informs residents of Stroud District Council when the next bin day is and what waste collection type it is.

This skill has been developed with the ![ASK CLI](https://developer.amazon.com/en-US/docs/alexa/smapi/quick-start-alexa-skills-kit-command-line-interface.html).
If you wish to deploy this skill to your Amazon developer account/AWS account, you can do this by running:
```
ask deploy
```
in the repo's top directory.

### N.B.
For security purposes I have sanitised the ![skills.json](https://github.com/jgcrunden/stroud-bin-day/blob/master/skill-package/skill.json) and the ![ask-states.json](https://github.com/jgcrunden/stroud-bin-day/blob/master/.ask/ask-states.json), removing references to things like the lambda ARN, S3 bucket ARN, skill ID etc, and replacing them with placeholders.
For this repo's CI/CD, I have added a simple sed command to swap the placeholders for their actual values, taking them from GitHub secret environment variables.
If you are deploying this skill to your own account, you will need to remove the placeholders before running the deploy command.

### N.B.
The ASK CLI will deploy the zipped copy of the lambda function to AWS Lambda via S3. During the zipping process the execute permissions of the go binary are lost, making in unexcutable. I've raised a bug on the Alexa ASK GitHub repo for this to be addressed, but in the meantime it is good practice to separately upload the zipped lambda after running an 'ask deploy'.
I have written a make target which uses the aws cli to deploy just the lambda zip. The can be run with the following
```
cd lambda
export ARN=<YOUR_LAMBDAS_ARN>
make deploy
```
Ensure that you are logged into your AWS account via the CLI before running this command.
