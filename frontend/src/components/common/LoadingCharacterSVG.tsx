import React from "react";

interface LoadingLemonSVGProps {
    message?: string;
    subMessage?: string;
}

const LoadingLemonSVG: React.FC<LoadingLemonSVGProps> = ({
                                                             message = "Loading...",
                                                             subMessage = "Setting up your DB"
                                                         }) => (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
        <rect width={200} height={200} fill="transparent" opacity={0}/>
        <g transform="translate(100, 100) scale(1.8)">
            <ellipse
                cx={0}
                cy={30}
                rx={45}
                ry={12}
                fill="#FFE0B2"
                stroke="#FF9800"
                strokeWidth={2}
            />
            <path
                d="M-45 30 L-45 0 C-45 -15, -25 -25, 0 -25 C25 -25, 45 -15, 45 0 L45 30"
                fill="#FFF3E0"
                stroke="#FF9800"
                strokeWidth={2}
            />
            <ellipse
                cx={0}
                cy={0}
                rx={45}
                ry={12}
                fill="#FFF8F0"
                stroke="#FF9800"
                strokeWidth={2}
            />
            <circle cx={-15} cy={3} r={4} fill="#333"/>
            <circle cx={15} cy={3} r={4} fill="#333"/>
            <path d="M-8 15 L8 15" stroke="#333" strokeWidth={3} fill="none"/>
            <circle
                cx={0}
                cy={8}
                r={12}
                fill="none"
                stroke="#FF9800"
                strokeWidth={2}
                strokeDasharray="15,10"
            >
                <animateTransform
                    attributeName="transform"
                    type="rotate"
                    values="0;360"
                    dur="2s"
                    repeatCount="indefinite"
                />
            </circle>
            <ellipse cx={0} cy={-32} rx={25} ry={20} fill="#FFEB3B"/>
            <ellipse cx={0} cy={-32} rx={20} ry={15} fill="#FFF59D"/>
            <path
                d="M-5 -48 L5 -48 L5 -44 L0 -40 L-5 -44 Z"
                fill="#4CAF50"
                stroke="#388E3C"
                strokeWidth={1}
            />
            <path
                d="M-40 8 Q-35 15, -30 10"
                stroke="#FF9800"
                strokeWidth={2}
                fill="none"
            />
            <path
                d="M40 8 Q35 15, 30 10"
                stroke="#FF9800"
                strokeWidth={2}
                fill="none"
            />
            <text
                x={0}
                y={65}
                fontFamily="Arial, sans-serif"
                fontSize={12}
                textAnchor="middle"
                fill="#333"
            >
                {message}
            </text>
            <text
                x={0}
                y={78}
                fontFamily="Arial, sans-serif"
                fontSize={10}
                textAnchor="middle"
                fill="#FF9800"
            >
                {subMessage}
            </text>
        </g>
    </svg>
);

export default LoadingLemonSVG;