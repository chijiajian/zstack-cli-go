using System;
using System.ComponentModel;
using System.Runtime.InteropServices;
using System.Text;
using Microsoft.Win32.SafeHandles;

namespace CloudContractTests
{
    public static class WindowsBenchmarkJob
    {
        private const int BasicAccountingInformation = 1;
        private const int ExtendedLimitInformation = 9;
        private const uint KillOnJobClose = 0x2000;
        private const uint ExtendedStartupInfoPresent = 0x00080000;
        private const uint CreateNoWindow = 0x08000000;
        private static readonly IntPtr JobListAttribute = new IntPtr(0x0002000D);

        public sealed record Result(uint ExitCode, ulong PeakCommitBytes, ulong LargestProcessPeakCommitBytes, uint ProcessCount);

        public static Result Run(string powerShellPath, string encodedCommand, uint timeoutMilliseconds)
        {
            using var job = CreateJobObjectW(IntPtr.Zero, null);
            if (job.IsInvalid) throw Failure("CreateJobObjectW");
            var limits = new ExtendedLimits { Basic = new BasicLimits { LimitFlags = KillOnJobClose } };
            uint limitsSize = (uint)Marshal.SizeOf<ExtendedLimits>();
            if (!SetInformationJobObject(job, ExtendedLimitInformation, ref limits, limitsSize))
                throw Failure("SetInformationJobObject");

            using var process = StartWorker(job, powerShellPath, encodedCommand);
            uint wait = WaitForSingleObject(process, timeoutMilliseconds);
            if (wait == 0x102) throw new TimeoutException("The query benchmark worker timed out.");
            if (wait != 0) throw Failure("WaitForSingleObject");
            if (!GetExitCodeProcess(process, out uint exitCode)) throw Failure("GetExitCodeProcess");
            if (!QueryLimits(job, ExtendedLimitInformation, out limits, limitsSize, IntPtr.Zero))
                throw Failure("QueryInformationJobObject (memory)");
            if (!QueryAccounting(job, BasicAccountingInformation, out var accounting, (uint)Marshal.SizeOf<Accounting>(), IntPtr.Zero))
                throw Failure("QueryInformationJobObject (process count)");
            return new Result(exitCode, limits.PeakJobMemoryUsed.ToUInt64(), limits.PeakProcessMemoryUsed.ToUInt64(), accounting.TotalProcesses);
        }

        private static SafeFileHandle StartWorker(SafeFileHandle job, string path, string encodedCommand)
        {
            // Assign the worker at creation so Windows counts its startup allocations and
            // inherits the same job for every CLI child. The fixture runs outside the job.
            UIntPtr listSize = UIntPtr.Zero;
            InitializeProcThreadAttributeList(IntPtr.Zero, 1, 0, ref listSize);
            if (listSize == UIntPtr.Zero) throw Failure("InitializeProcThreadAttributeList (size)");
            IntPtr attributes = Marshal.AllocHGlobal(checked((int)listSize.ToUInt64()));
            bool initialized = false;
            IntPtr jobList = IntPtr.Zero;
            try
            {
                initialized = InitializeProcThreadAttributeList(attributes, 1, 0, ref listSize);
                if (!initialized) throw Failure("InitializeProcThreadAttributeList");
                jobList = Marshal.AllocHGlobal(IntPtr.Size);
                Marshal.WriteIntPtr(jobList, job.DangerousGetHandle());
                if (!UpdateProcThreadAttribute(attributes, 0, JobListAttribute, jobList, (UIntPtr)IntPtr.Size, IntPtr.Zero, IntPtr.Zero))
                    throw Failure("UpdateProcThreadAttribute (job list)");
                var startup = new StartupInfoEx
                {
                    Info = new StartupInfo { Size = (uint)Marshal.SizeOf<StartupInfoEx>() },
                    Attributes = attributes
                };
                var command = new StringBuilder($"\"{path}\" -NoLogo -NoProfile -NonInteractive -EncodedCommand {encodedCommand}");
                if (!CreateProcessW(path, command, IntPtr.Zero, IntPtr.Zero, false, ExtendedStartupInfoPresent | CreateNoWindow,
                    IntPtr.Zero, null, ref startup, out var process))
                    throw Failure("CreateProcessW (benchmark worker)");
                CloseHandle(process.Thread);
                return new SafeFileHandle(process.Process, ownsHandle: true);
            }
            finally
            {
                if (initialized) DeleteProcThreadAttributeList(attributes);
                Marshal.FreeHGlobal(jobList);
                Marshal.FreeHGlobal(attributes);
            }
        }

        private static Win32Exception Failure(string operation)
        {
            int error = Marshal.GetLastWin32Error();
            return new Win32Exception(error, $"{operation} failed (Win32 error {error}).");
        }

        [StructLayout(LayoutKind.Sequential)]
        private struct BasicLimits
        {
            public long PerProcessUserTimeLimit, PerJobUserTimeLimit;
            public uint LimitFlags;
            public UIntPtr MinimumWorkingSetSize, MaximumWorkingSetSize;
            public uint ActiveProcessLimit;
            public UIntPtr Affinity;
            public uint PriorityClass, SchedulingClass;
        }

        [StructLayout(LayoutKind.Sequential)]
        private struct ExtendedLimits
        {
            public BasicLimits Basic;
            // IO_COUNTERS contains six consecutive 64-bit counts.
            public ulong ReadOperations, WriteOperations, OtherOperations, ReadBytes, WriteBytes, OtherBytes;
            public UIntPtr ProcessMemoryLimit, JobMemoryLimit, PeakProcessMemoryUsed, PeakJobMemoryUsed;
        }

        [StructLayout(LayoutKind.Sequential)]
        private struct Accounting
        {
            public long TotalUserTime, TotalKernelTime, PeriodUserTime, PeriodKernelTime;
            public uint TotalPageFaultCount, TotalProcesses, ActiveProcesses, TerminatedProcesses;
        }

        [StructLayout(LayoutKind.Sequential)]
        private struct StartupInfo
        {
            public uint Size;
            public IntPtr Reserved, Desktop, Title;
            public uint X, Y, XSize, YSize, XCountChars, YCountChars, FillAttribute, Flags;
            public ushort ShowWindow, ReservedSize;
            public IntPtr ReservedData, StandardInput, StandardOutput, StandardError;
        }

        [StructLayout(LayoutKind.Sequential)]
        private struct StartupInfoEx
        {
            public StartupInfo Info;
            public IntPtr Attributes;
        }

        [StructLayout(LayoutKind.Sequential)]
        private struct ProcessInformation
        {
            public IntPtr Process, Thread;
            public uint ProcessId, ThreadId;
        }

        [DllImport("kernel32.dll", CharSet = CharSet.Unicode, ExactSpelling = true, SetLastError = true)]
        private static extern SafeFileHandle CreateJobObjectW(IntPtr securityAttributes, string name);
        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern bool SetInformationJobObject(SafeFileHandle job, int infoClass, ref ExtendedLimits info, uint size);
        [DllImport("kernel32.dll", EntryPoint = "QueryInformationJobObject", SetLastError = true)]
        private static extern bool QueryLimits(SafeFileHandle job, int infoClass, out ExtendedLimits info, uint size, IntPtr returnSize);
        [DllImport("kernel32.dll", EntryPoint = "QueryInformationJobObject", SetLastError = true)]
        private static extern bool QueryAccounting(SafeFileHandle job, int infoClass, out Accounting info, uint size, IntPtr returnSize);
        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern bool InitializeProcThreadAttributeList(IntPtr list, int count, uint flags, ref UIntPtr size);
        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern bool UpdateProcThreadAttribute(IntPtr list, uint flags, IntPtr attribute, IntPtr value, UIntPtr size, IntPtr previous, IntPtr returnSize);
        [DllImport("kernel32.dll")]
        private static extern void DeleteProcThreadAttributeList(IntPtr list);
        [DllImport("kernel32.dll", CharSet = CharSet.Unicode, ExactSpelling = true, SetLastError = true)]
        private static extern bool CreateProcessW(string application, StringBuilder command, IntPtr processAttributes, IntPtr threadAttributes,
            bool inheritHandles, uint flags, IntPtr environment, string currentDirectory, ref StartupInfoEx startup, out ProcessInformation process);
        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern uint WaitForSingleObject(SafeFileHandle handle, uint milliseconds);
        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern bool GetExitCodeProcess(SafeFileHandle process, out uint exitCode);
        [DllImport("kernel32.dll")]
        private static extern bool CloseHandle(IntPtr handle);
    }
}
