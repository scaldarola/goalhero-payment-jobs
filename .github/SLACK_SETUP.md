# Slack Notifications Setup

This repository is configured to send Slack notifications when tests fail or succeed. Follow these steps to set up Slack integration.

## 🔧 Setup Instructions

### 1. Create a Slack Webhook URL

1. Go to your Slack workspace settings
2. Navigate to **Apps** → **Custom Integrations** → **Incoming Webhooks**
3. Click **Add to Slack**
4. Choose the default channel (you can override this in workflows)
5. Copy the **Webhook URL** (it will look like `https://hooks.slack.com/services/...`)

### 2. Add the Webhook URL to GitHub Secrets

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Name: `SLACK_WEBHOOK_URL`
5. Value: Paste your Slack webhook URL
6. Click **Add secret**

### 3. Recommended Slack Channels

The workflows are configured to post to different channels based on the type of notification:

- `#deployments` - Success notifications
- `#alerts` - Test and build failures

**Create these channels in your Slack workspace or modify the channel names in the workflow files.**

## 📢 Notification Types

### ✅ Success Notifications
Sent to `#deployments` when:
- All tests pass
- Build succeeds

### ❌ Failure Notifications
Sent to `#alerts` when:
- **Tests fail**
- **Build fails**

## 🎨 Message Format

Success messages include:
- ✅ Status indicator
- 📝 Branch and commit info
- 👤 Author information
- 📊 Summary of passed checks (Tests, Build)
- 🔗 Link to view pipeline details

Failure messages include:
- ❌ Failure indicator
- 📝 Branch and commit info  
- 👤 Author information
- 🔍 Specific failure type (Tests or Build)
- 🔗 Direct link to logs

## 🔧 Customizing Notifications

To modify notifications, edit these files:
- `.github/workflows/tests-on-main.yml` - Main branch test notifications
- `.github/workflows/ci.yml` - Full CI pipeline notifications
- `.github/workflows/test.yml` - Basic test workflow notifications

### Example: Change notification channel
```yaml
channel: '#your-custom-channel'
```

### Example: Modify message format
```yaml
text: |
  Your custom message here
  *Branch:* `${{ github.ref_name }}`
  *Status:* ${{ job.status }}
```

## 🧪 Testing Notifications

1. Make sure `SLACK_WEBHOOK_URL` secret is set
2. Push code to main branch
3. Check your configured Slack channels for notifications

## 🚨 Troubleshooting

### No notifications received?
1. Verify `SLACK_WEBHOOK_URL` secret exists and is correct
2. Check that Slack channels exist
3. Ensure webhook has permission to post to channels
4. Check GitHub Actions logs for Slack notification steps

### Notifications going to wrong channel?
- The webhook has a default channel, but workflows can override it
- Verify channel names in workflow files match your Slack channels

### Want to disable notifications?
Comment out or remove the Slack notification steps in the workflow files.

## 🔐 Security Notes

- The `SLACK_WEBHOOK_URL` is stored as a GitHub secret and is not visible in logs
- Only repository collaborators can view/edit secrets
- Webhook URLs should be kept confidential