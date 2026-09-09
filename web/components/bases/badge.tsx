import type { ComponentProps } from "react";

import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex shrink-0 items-center justify-center rounded-full wongnok-text-label font-bold whitespace-nowrap",
  {
    variants: {
      variant: {
        contained: "border border-transparent",
        outlined: "border",
      },
      color: {
        primary: "",
        accent: "",
        success: "",
        error: "",
        gray: "",
      },
      size: {
        medium: "px-2.5 py-0.5",
        large: "px-3.5 py-1",
      },
    },
    defaultVariants: {
      variant: "contained",
      color: "primary",
      size: "medium",
    },
    compoundVariants: [
      {
        variant: "contained",
        color: "primary",
        className: "bg-primary text-primary-foreground",
      },
      {
        variant: "contained",
        color: "accent",
        className: "bg-accent text-accent-foreground",
      },
      {
        variant: "contained",
        color: "success",
        className: "bg-success text-primary-foreground",
      },
      {
        variant: "contained",
        color: "error",
        className: "bg-destructive text-primary-foreground",
      },
      {
        variant: "contained",
        color: "gray",
        className: "bg-muted-foreground text-primary-foreground",
      },
      {
        variant: "outlined",
        color: "primary",
        className: "border-primary text-primary",
      },
      {
        variant: "outlined",
        color: "accent",
        className: "border-accent text-accent",
      },
      {
        variant: "outlined",
        color: "success",
        className: "border-success text-success",
      },
      {
        variant: "outlined",
        color: "error",
        className: "border-destructive text-destructive",
      },
      {
        variant: "outlined",
        color: "gray",
        className: "border-muted-foreground text-muted-foreground",
      },
    ],
  },
);

export type BadgeProps = ComponentProps<"span"> &
  VariantProps<typeof badgeVariants>;

function Badge({
  className,
  variant = "contained",
  size = "medium",
  color = "primary",
  ...props
}: BadgeProps) {
  return (
    <span
      data-slot="badge"
      className={cn(badgeVariants({ variant, size, color, className }))}
      {...props}
    />
  );
}

export default Badge;
export { Badge, badgeVariants };
