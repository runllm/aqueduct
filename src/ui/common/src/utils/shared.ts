export const ContentSidebarOffsetInPx = 100;

export enum LoadingStatusEnum {
  Initial = 'initial',
  Loading = 'loading',
  Failed = 'failed',
  Succeeded = 'succeeded',
}

export type LoadingStatus = {
  loading: LoadingStatusEnum;
  err: string;
};

export function isInitial(status: LoadingStatus) {
  return status.loading === LoadingStatusEnum.Initial;
}

export function isLoading(status: LoadingStatus) {
  return status.loading === LoadingStatusEnum.Loading;
}

export function isSucceeded(status: LoadingStatus) {
  return status.loading === LoadingStatusEnum.Succeeded;
}

export function isFailed(status: LoadingStatus) {
  return status.loading === LoadingStatusEnum.Failed;
}

export enum ExecutionStatus {
  Unknown = 'unknown',
  Succeeded = 'succeeded',
  Failed = 'failed',
  Pending = 'pending',
  Canceled = 'canceled',
  Registered = 'registered',
}

export type ExecState = {
  status: ExecutionStatus;
  failure_type?: FailureType;
  error?: Error;
  user_logs?: Logs;
};

export enum FailureType {
  System = 1,
  UserFatal = 2,
  UserNonFatal = 3,
}

export enum CheckStatus {
  Succeeded = 'True',
  Failed = 'False',
}

export default ExecutionStatus;
export const TransitionLengthInMs = 200;

export const WidthTransition = `width ${TransitionLengthInMs}ms ease-in-out`;
export const HeightTransition = `height ${TransitionLengthInMs}ms ease-in-out`;
export const AllTransition = `all ${TransitionLengthInMs}ms ease-in-out`;

export type Logs = {
  stdout?: string;
  stderr?: string;
};

export type Error = {
  context?: string;
  tip?: string;
};

export const GithubIssueLink = `https://github.com/aqueducthq/aqueduct/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D`;
