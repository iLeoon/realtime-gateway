import { z } from 'zod';

export const ErrorCodeSchema = z.enum([
  'BadRequest',
  'ForbiddenRequest',
  'UnauthorizedRequest',
  'NotFoundRequest',
  'InternalServerError',
  'BadArgumet',
  'BadGateway',
  'Timeout',
  'ServiceUnavailable',
  'StatusNotAccepted'
]);

export const ErrorDetailsSchema = z
  .object({
    code: z.string().min(1),
    target: z.string().min(1),
    message: z.string().min(1)
  })
  .strict();

export const InnerErrorDetailsSchema = z
  .object({
    code: z.string().min(1)
  })
  .strict();

export const InnerErrorSchema = z
  .object({
    code: z.string().min(1),
    innererror: InnerErrorDetailsSchema.optional()
  })
  .strict();

export const ErrorBodySchema = z
  .object({
    code: ErrorCodeSchema,
    message: z.string().min(1),
    target: z.string().min(1).optional(),
    details: z.array(ErrorDetailsSchema).optional(),
    innererror: InnerErrorSchema.optional()
  })
  .strict();

export const ErrorResponseSchema = z
  .object({
    error: ErrorBodySchema
  })
  .strict();

export type ErrorResponse = z.infer<typeof ErrorResponseSchema>;
