function joinClasses(...values: Array<string | undefined>) {
  return values.filter(Boolean).join(" ")
}

export function BriefcastLogo({
  size = 24,
  className,
}: {
  size?: number
  className?: string
}) {
  return (
    <img
      src="/brand-logo-placeholder.svg"
      alt=""
      width={size}
      height={size}
      className={joinClasses("shrink-0", className)}
      aria-hidden="true"
    />
  )
}

export function BriefcastWordmark({
  className,
  accentClassName,
}: {
  className?: string
  accentClassName?: string
}) {
  return (
    <span
      className={joinClasses("font-display text-base font-semibold tracking-[-0.03em] text-foreground", className)}
    >
      Brief<span className={accentClassName ?? "text-primary"}>cast</span>
    </span>
  )
}
