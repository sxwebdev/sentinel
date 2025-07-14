export const ActivityIndicatorSVG = ({active = true, size = 16}) => {
  const color = active ? "#3b82f6" : "#ef4444";
  const pulseColor = active ? "#3b82f6" : "#ef4444";

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 32 32"
      style={{display: "inline-block", verticalAlign: "middle"}}
    >
      {/* Пульсирующее кольцо */}
      <circle cx="16" cy="16" r="8" fill={pulseColor} opacity="0.3">
        <animate
          attributeName="r"
          values="8;15"
          dur="1.3s"
          repeatCount="indefinite"
        />
        <animate
          attributeName="opacity"
          values="0.8;0.3"
          dur="1.3s"
          repeatCount="indefinite"
        />
      </circle>
      {/* Статичный центр */}
      <circle cx="16" cy="16" r="8" fill={color} />
    </svg>
  );
};
