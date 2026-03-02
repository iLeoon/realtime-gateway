type LogoutHandler = () => void;

let logoutHandler: LogoutHandler | null = null;

export const registerLogoutHandler = (handler: LogoutHandler): void => {
  logoutHandler = handler;
};

export const triggerLogout = (): void => {
  if (logoutHandler) {
    logoutHandler();
  }
};
