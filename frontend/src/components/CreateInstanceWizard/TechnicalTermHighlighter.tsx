import React, {useState, useRef, useEffect} from 'react';
import ReactDOM from 'react-dom';
import './TechnicalTermHighlighter.css';

interface TechnicalTermHighlighterProps {
    text: string;
    terms: Record<string, string>;
    className?: string;
}

interface TermPopoverProps {
    term: string;
    definition: string;
    targetRect: DOMRect;
    onClose: () => void;
}

const TermPopover: React.FC<TermPopoverProps> = ({term, definition, targetRect, onClose}) => {
    const popoverRef = useRef<HTMLDivElement>(null);
    const [position, setPosition] = useState({top: 0, left: 0});
    const [isAbove, setIsAbove] = useState(false);
    const [isVisible, setIsVisible] = useState(false);

    useEffect(() => {
        // 팝오버가 렌더링된 후 위치 계산
        const calculatePosition = () => {
            if (!popoverRef.current) return;

            const popover = popoverRef.current;
            const popoverRect = popover.getBoundingClientRect();

            // 뷰포트 크기
            const viewportWidth = window.innerWidth;
            const viewportHeight = window.innerHeight;

            // 팝오버 크기
            const popoverWidth = popoverRect.width || 300; // 기본값 설정
            const popoverHeight = popoverRect.height || 100; // 기본값 설정

            // 기본 위치: 타겟 요소의 중앙 아래
            let left = targetRect.left + (targetRect.width / 2) - (popoverWidth / 2);
            let top = targetRect.bottom + 10; // 10px 간격

            // 좌우 경계 체크
            if (left < 10) {
                left = 10;
            } else if (left + popoverWidth > viewportWidth - 10) {
                left = viewportWidth - popoverWidth - 10;
            }

            // 상하 위치 결정
            let showAbove = false;

            // 아래 공간이 부족한 경우
            if (top + popoverHeight > viewportHeight - 10) {
                // 위에 표시할 공간이 있는지 확인
                const topSpace = targetRect.top - 10;
                if (topSpace > popoverHeight) {
                    // 위에 표시
                    top = targetRect.top - popoverHeight - 10;
                    showAbove = true;
                } else {
                    // 위아래 모두 공간이 부족하면 가능한 위치에 표시
                    if (topSpace > viewportHeight - targetRect.bottom) {
                        // 위쪽이 더 넓으면 위에
                        top = Math.max(10, targetRect.top - popoverHeight - 10);
                        showAbove = true;
                    } else {
                        // 아래쪽이 더 넓으면 아래에
                        top = Math.min(targetRect.bottom + 10, viewportHeight - popoverHeight - 10);
                    }
                }
            }

            setPosition({top, left});
            setIsAbove(showAbove);
            setIsVisible(true);
        };

        // DOM이 업데이트된 후 위치 계산
        requestAnimationFrame(() => {
            calculatePosition();
        });

        // 스크롤이나 리사이즈 시 팝오버 닫기
        const handleScrollOrResize = () => {
            onClose();
        };

        window.addEventListener('scroll', handleScrollOrResize, true);
        window.addEventListener('resize', handleScrollOrResize);

        return () => {
            window.removeEventListener('scroll', handleScrollOrResize, true);
            window.removeEventListener('resize', handleScrollOrResize);
        };
    }, [targetRect, onClose]);

    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            if (popoverRef.current && !popoverRef.current.contains(e.target as Node)) {
                onClose();
            }
        };

        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                onClose();
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        document.addEventListener('keydown', handleEscape);

        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
            document.removeEventListener('keydown', handleEscape);
        };
    }, [onClose]);

    return (
        <div
            ref={popoverRef}
            className={`term-popover ${isAbove ? 'above' : 'below'} ${isVisible ? 'visible' : ''}`}
            style={{
                top: `${position.top}px`,
                left: `${position.left}px`
            }}
        >
            <div className="popover-arrow"></div>
            <div className="popover-header">
                <span className="popover-term">{term}</span>
                <button className="popover-close" onClick={onClose}>×</button>
            </div>
            <div className="popover-content">{definition}</div>
        </div>
    );
};

const TechnicalTermHighlighter: React.FC<TechnicalTermHighlighterProps> = ({
                                                                               text,
                                                                               terms,
                                                                               className = ''
                                                                           }) => {
    const [activePopover, setActivePopover] = useState<{
        term: string;
        definition: string;
        rect: DOMRect;
    } | null>(null);

    // 기술 용어를 찾아서 처리
    const processText = () => {
        if (!text || !terms || Object.keys(terms).length === 0) {
            return <span>{text}</span>;
        }

        // 모든 용어를 정규식으로 만들기 (긴 용어부터 매칭하도록 정렬)
        const sortedTerms = Object.keys(terms).sort((a, b) => b.length - a.length);
        const pattern = sortedTerms.map(term =>
            term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') // 특수문자 이스케이프
        ).join('|');

        const regex = new RegExp(`(${pattern})`, 'gi');
        const parts = text.split(regex);

        return parts.map((part, index) => {
            const matchedTerm = sortedTerms.find(term =>
                term.toLowerCase() === part.toLowerCase()
            );

            if (matchedTerm) {
                return (
                    <span
                        key={index}
                        className="technical-term"
                        onClick={(e) => {
                            e.stopPropagation();
                            const rect = (e.target as HTMLElement).getBoundingClientRect();

                            // 이미 같은 요소의 팝오버가 열려있으면 닫기
                            if (activePopover &&
                                activePopover.term === matchedTerm &&
                                activePopover.rect.top === rect.top &&
                                activePopover.rect.left === rect.left) {
                                setActivePopover(null);
                            } else {
                                setActivePopover({
                                    term: matchedTerm,
                                    definition: terms[matchedTerm],
                                    rect
                                });
                            }
                        }}
                    >
                        {part}
                    </span>
                );
            }
            return <span key={index}>{part}</span>;
        });
    };

    // Portal container 생성
    useEffect(() => {
        let portalContainer = document.getElementById('term-popover-portal');
        if (!portalContainer) {
            portalContainer = document.createElement('div');
            portalContainer.id = 'term-popover-portal';
            document.body.appendChild(portalContainer);
        }

        return () => {
            // 컴포넌트 언마운트 시 activePopover 정리
            setActivePopover(null);
        };
    }, []);

    return (
        <>
            <span className={className}>
                {processText()}
            </span>
            {activePopover &&
                ReactDOM.createPortal(
                    <TermPopover
                        term={activePopover.term}
                        definition={activePopover.definition}
                        targetRect={activePopover.rect}
                        onClose={() => setActivePopover(null)}
                    />,
                    document.body // body에 직접 렌더링
                )
            }
        </>
    );
};

export default TechnicalTermHighlighter;