let token: string | null = null;

export const tokenStore = {
  getToken: (): string | null => token,
  setToken: (nextToken: string | null): void => {
    token = nextToken;
  },
  clearToken: (): void => {
    token = null;
  }
};
