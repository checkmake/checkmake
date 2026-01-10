# The checkmake release process

## release pattern

checkmake releases are following the version  pattern vX.Y.Z.
Here, X is the *major version*, Y is the *minor version*, and Z is the *patch version*.

The checkmake project is  roughly following the [semantic versioning](https://semver.org/) pattern, i. e. 

* X is bumped when  adding major, breaking changes
* Y is bumped when  adding new features
* Z is bumped when adding only bugfixes

## Release Principles

Major and minor releases are released directly from the `main` branch.

Upon creation of a minor release vX.Y, a branch `release-X.Y` is created with its head at the release
and patch releases vX.Y.Z will be created from the branch `release-X.Y` 

The following instructions for creating releases require push rights (i. e. maintainer rights) to the checkmake repository.


## Creating A Major Or Minor Release

This section explains release creation at the example of v0.3.0.

Generalizing it to other releases should be straightforward.




### 1. prepare release notes

Before the release  can be created, the release notes need to be prepared in  the main branch the file in [docs/releases](https://github.com/checkmake/checkmake/tree/main/docs/releases),
in this example [v0.3.0.md](https://github.com/checkmake/checkmake/blob/main/docs/releases/v0.3.0.md).

It is important that the releaso notes file's base name is exactly the version tpo be created, i. e. in this case `v0.3.0`. 

Existing release note files can be used as blueprints.

this release notes file  will usually be added  with a pull request  but maintainers can also push directly.

### 2. check out the main branch

Once the release-notes file exists in the main branch, check out the current `main` branch  locally:

```console
$ cd $CHECKMAKE_GIT_CHECKOUT_DIR
$ git checkout main 
$ git fetch origin
$ dit reset --hard origin/main
```

This assumes that the local chgeckout has a remote configured pointing to the main checkmake repo on github.

The top (`HEAD`) commit should now be the commit that adds the release notes file.

### 3. create and push the release tag

Now tag the HEAD and push the tag:

```console
$ git tag -a -m "release v0.3.0" v0.3.0
$ git push origin main
```

### 4. create the github release


This step  requires the [github gh cli tool](https://cli.github.com/) to be installed locally.

```console
$ make github-release
```

This creates [the release on github](https://github.com/checkmake/checkmake/releases/tag/v0.3.0) including build and upload of artifacts.

### 5. release branch

Next, create and push the release branch:

```console
$ git checkout -b release-0.3 v0.3.0
$ git push origin release-0.3
```

### 6. container images

Finally, create and publish container images.

```console
$ make IMAGE_VERSION_TAG=v0.3.0 image-build
$podman login quay.io
$ make IMAGE_VERSION_TAG=v0.3.0 image-push
```
Use docker instead of podman if that is your primary container command.

make will auto-detect the tool with a preference to podman.

## Creating A Patch Release

These instruction explain the patch release proces at the example of v0.3.2.

## 1. Backport Bugfixes

First, all required bugfixes need to be backported from main to the release branch release-0.3.

This is typicaly done with a `git cherry-pick`on the release branch, e. e.:

```console
$ git checkout release-0.3
$ git fetch origin
$ git reset --hard origin/release-0.3
$ git cherry-pick -x COMMIT_SHA_FROM_MAIN
```

When backporting a pull request with multiplew commits, the cherry-pick needs to be done for every commit from that PR.

Create a branch from the new HEAD of the local release-0.3 branch und use this to open a PR against the release-0.3 brancxh on github.


## 2. Release Notes

The nexct step is to prepare release notes.

These are created in main (e. g. PR #236) and then backported to release-0.3 (e. g. PR #238)

## 3. Tag the release and push

Make sure the local release branch is up to date with the release notes and tag the release notes commit:

```console
$ git checkout release-0.3
$ git fetch origin
$ git reset --hard origin/release-0.3
$ git tag -a  v0.3.2 -m "release v0.3.2"
$ git push origin --tags
```

## 4. Create  the github release 

```console
$ make github-release
```

Note that this requires the github gh cli tool to be installed locally and logged in (`gh auth login`).

$$ 5. build and push container images

```console
$ podman login quay.io
$ make IMAGE_VERSION_TAG=v0.3.2 image-build
$ make IMAGE_VERSION_TAG=v0.3.2 image-push
```

































