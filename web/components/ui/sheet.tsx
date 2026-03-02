'use client';

import * as React from 'react';
import { createPortal } from 'react-dom';

import { cn } from '@/lib/utils';

type SheetContextValue = {
  open: boolean;
  setOpen: (open: boolean) => void;
};

const SheetContext = React.createContext<SheetContextValue | null>(null);

const useSheetContext = () => {
  const context = React.useContext(SheetContext);
  if (!context) {
    throw new Error('Sheet components must be used inside <Sheet>.');
  }

  return context;
};

function Sheet({ open, onOpenChange, children }: { open: boolean; onOpenChange: (open: boolean) => void; children: React.ReactNode }) {
  return <SheetContext.Provider value={{ open, setOpen: onOpenChange }}>{children}</SheetContext.Provider>;
}

function SheetTrigger({ children }: { children: React.ReactElement<{ onClick?: () => void }> }) {
  const { setOpen } = useSheetContext();

  return React.cloneElement(children, {
    onClick: () => setOpen(true)
  });
}

function SheetClose({ children }: { children: React.ReactElement<{ onClick?: () => void }> }) {
  const { setOpen } = useSheetContext();

  return React.cloneElement(children, {
    onClick: () => setOpen(false)
  });
}

function SheetPortal({ children }: { children: React.ReactNode }) {
  const [mounted, setMounted] = React.useState(false);

  React.useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return null;
  }

  return createPortal(children, document.body);
}

function SheetOverlay({ className }: { className?: string }) {
  const { open, setOpen } = useSheetContext();

  if (!open) {
    return null;
  }

  return (
    <SheetPortal>
      <button
        type="button"
        aria-label="Close sidebar"
        onClick={() => setOpen(false)}
        className={cn('fixed inset-0 z-40 bg-black/60 backdrop-blur-[2px] transition-opacity duration-200', className)}
      />
    </SheetPortal>
  );
}

function SheetContent({
  children,
  className,
  side = 'left'
}: {
  children: React.ReactNode;
  className?: string;
  side?: 'left' | 'right';
}) {
  const { open } = useSheetContext();

  if (!open) {
    return null;
  }

  return (
    <SheetPortal>
      <div
        className={cn(
          'fixed top-0 z-50 h-full w-[88vw] max-w-sm border-border/70 bg-background/95 p-0 shadow-2xl backdrop-blur-md transition-transform duration-200',
          side === 'left' ? 'left-0 border-r' : 'right-0 border-l',
          className
        )}
      >
        {children}
      </div>
    </SheetPortal>
  );
}

function SheetHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('flex items-center justify-between px-4 py-3', className)} {...props} />;
}

function SheetTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return <h2 className={cn('text-sm font-semibold tracking-tight', className)} {...props} />;
}

export { Sheet, SheetClose, SheetContent, SheetHeader, SheetOverlay, SheetTitle, SheetTrigger };
