# Acceptance evidence template

Record the module version, CLI version and revision, package SHA-256, OEM brand, OS and architecture, PowerShell version, and Pester version. List the exact commands tested and any skipped conditions.

For the 10,000-VM benchmark, record the latency percentiles, pagination count and `peakCombinedCommitBytes`. On Windows, this is the peak concurrent commit charge of the isolated benchmark PowerShell process and its CLI children, including warmup and cleanup. The harness and fixture run outside that job. Record the 512 MiB target result separately. On other platforms the combined peak and target result are reported as unavailable.

Separate local mock evidence, Windows native results, authorized ZStack tests, and customer script migration results. For a customer sample, record whether the command replacement was sufficient, the parameters or object fields that changed, and the unsupported behavior. Keep scripts and logs sanitized.

For live testing, record authorization, resource UUID, initial state, requested actions, final observed state, and cleanup. On an unknown write outcome, check the resource before retrying. Keep credentials and raw server responses outside the record.
