export type AppError = {
  code: string;
  message: string;
  target?: string;
  innerCode?: string;
  status?: number;
};

export class SchemaValidationError extends Error {
  readonly issues: string[];

  constructor(entity: string, issues: string[]) {
    super(`Invalid ${entity} response schema`);
    this.name = 'SchemaValidationError';
    this.issues = issues;
  }
}
