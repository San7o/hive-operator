# Testing

Currently, the operator will log inodes of the container's processes
of kubernete's Pods. After deployment, wou should see the information
logged in the operator's standard output.

When testing, It may be useful to kill all the pods, which can be
done with:
```bash
make kill-pods-local
```
