export function BriefcastLogo({ size = 24, className }: { size?: number; className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 40 40"
      fill="none"
      width={size}
      height={size}
      className={className}
      aria-hidden="true"
    >
      <path d="M12 8v20c0 1 1 2 2 2h1V10c0-1-1-2-2-2h-1z" fill="currentColor" />
      <path d="M17 6v24c0 .5.5 1 1 1s1-.5 1-1V6c0-.5-.5-1-1-1s-1 .5-1 1z" fill="currentColor" />
      <path d="M21 4v28c0 .5.5 1 1 1s1-.5 1-1V4c0-.5-.5-1-1-1s-1 .5-1 1z" fill="currentColor" />
      <path d="M25 6v24c0 .5.5 1 1 1s1-.5 1-1V6c0-.5-.5-1-1-1s-1 .5-1 1z" fill="currentColor" />
      <path d="M29 8v20c0 1-1 2-2 2h-1V10c0-1 1-2 2-2h1z" fill="currentColor" />
      <path d="M12 28c0 2 4 4 8 4s8-2 8-4" stroke="currentColor" strokeWidth="1.5" fill="none" strokeLinecap="round" />
    </svg>
  )
}
