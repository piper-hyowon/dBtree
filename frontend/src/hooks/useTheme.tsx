import React, { createContext, useContext, useEffect, useState } from "react";

// TODO: Context + Provider + Hook 이 한 파일에 있음
// 필요시 src/contexts/ 만들어서 ThemeContext.tsx 로 분리

type Theme = "light" | "dark";

interface ThemeContextType {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  isNight: boolean;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({
                                                                           children,
                                                                       }) => {
    // localStorage에서 저장된 테마를 가져오거나 기본값으로 'light' 사용
    const [theme, setThemeState] = useState<Theme>(() => {
        const savedTheme = localStorage.getItem('theme');
        return (savedTheme as Theme) || 'light';
    });

    const isNight = theme === "dark";

    // 테마 변경 시 localStorage에도 저장
    const setTheme = (newTheme: Theme) => {
        setThemeState(newTheme);
        localStorage.setItem('theme', newTheme);
    };

    useEffect(() => {
        document.body.dataset.theme = theme;
    }, [theme]);

    return (
        <ThemeContext.Provider value={{ theme, setTheme, isNight }}>
            {children}
        </ThemeContext.Provider>
    );
};

export const useTheme = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }
  return context;
};
