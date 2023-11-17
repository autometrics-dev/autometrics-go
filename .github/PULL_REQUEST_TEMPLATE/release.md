# Release <Version Number>

- [ ] The `CHANGELOG` is updated with a new section and the correct links
- [ ] The `Version` value in `internal/build` package is updated

Once this PR is merged, the commit must have a matching tag in the repository, 
and a corresponding release must be created in Github. If no other PR is merged
after this one, it is possible to do both steps in one go by creating the release
from Github UI. Otherwise, put the tag on the correct commit first, and do
the release "from an existing tag".
