# k8s-crash-informer
[![Build](https://github.com/lnsp/k8s-crash-informer/workflows/Go/badge.svg)](https://github.com/lnsp/k8s-crash-informer/actions?workflow=Go) ![Docker Cloud](https://img.shields.io/docker/cloud/build/lnsp/crash-informer)

This Kubernetes controller informs you when a Kubernetes Pod repeatedly dies (`CrashLoopBackOff`) while providing additional information like exit code and logs. **This is my first attempt at writing a Kubernetes controller, if you have any feedback please open an issue.**

## Usage

### Step 1: Add a config map for the informer
##### For Mattermost, you should use
```yaml
apiVersion: v1
data:
  token: <bot-token>
  channel: <channel-name>
  team: <team-name>
  url: <your-mattermost-url>
kind: ConfigMap
metadata:
  name: mattermost-informer-cfg
```

You should supply the [token of the bot](https://docs.mattermost.com/developer/bot-accounts.html) (or a [personal access token of a user](https://docs.mattermost.com/developer/personal-access-tokens.html)), server URL, team and channel.

##### If you use Slack, use
```yaml
apiVersion: v1
data:
  channel: <channel-name>
  token: <your-token>
kind: ConfigMap
metadata:
  name: slack-informer-cfg
```

You should use the Bot User OAuth Access Token as `token`. It can be copied from the Slack App admin interface after registering a new Slack API App and enabling the Bot feature.

This step is required to create a valid configuration for our crash informer.

### Step 2: Deploy the informer
```bash
# If you use Mattermost
kubectl apply -f manifests/mattermost-informer.yaml

# If you use Slack
kubectl apply -f manifests/slack-informer.yaml
```

You may want to update the `namespace` references, since the informer only watches a given namespace.

### Step 3: Annotate your Deployments, ReplicaSets or Pods
To begin watching pods (or deployments or replica sets), you only have to add the following annotation to the spec.

```yaml
annotations:
  espe.tech/crash-informer: "true"
```

You may optionally set the backoff interval in seconds using `espe.tech/informer-backoff`.

### Step 4: Get notified!

![Notification in Mattermost](https://i.imgur.com/BzJnaRr.png)
