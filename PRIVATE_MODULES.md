# Usage of private modules set up

Sometimes you have to work with private modules in your Go application. For example, you may
need to use some common libraries that are not open-source, or you may need to use some
internal libraries that are not ready to be published yet. This document describes how to set
up your Go environment to work with private modules.

## Setting up Go environment

Let's assume you have a private module hosted on Github and it is: `github.com/some-secret-repo`

First, you need to tell Go that you want to access private repositories. You can do this by

```shell
go env -w GOPRIVATE=github.com/some-secret-repo
```

At this point ability to access private repositories is added, but no credentials are
provided yet. To tell Git how to log in on your behalf - use the `~/.netrc` file.
`.netrc` file contains various host names as well as login credentials for those hosts.
To give Git your credentials, you’ll need to have a `.netrc` that includes github.com in your
home directory.

First you need to create `~/.netrc` file (check if it exists already) and add your credentials
to it. It should look like this:

```
machine github.com
login your_github_username
password your_github_access_token
```

replace **your_github_username** with your Github username and **your_github_access_token** with
token generated here: https://github.com/settings/tokens

Make sure to select **Tokens (class)** and then click **Generate new token**. You can give it
any expiration time you want and select all options in **repo** section. Once you have your
token, you can use it as a password in `.netrc` file.
