import React from "react";
import { useTheme } from "../../../hooks/useTheme";
import "./ToggleThemeButton.css";

const ToggleThemeButton: React.FC = () => {
  const { theme, setTheme } = useTheme();

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  return (
    <button
      className="theme-toggle-button"
      onClick={toggleTheme}
      aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
    >
      <div className="icon-container">
        {theme === "dark" ? (
          <div className="sun-icon">
            <div className="sun-inner"></div>
            <div className="ray ray1"></div>
            <div className="ray ray2"></div>
            <div className="ray ray3"></div>
            <div className="ray ray4"></div>
            <div className="ray ray5"></div>
            <div className="ray ray6"></div>
            <div className="ray ray7"></div>
            <div className="ray ray8"></div>
          </div>
        ) : (
          <div className="moon-icon">
            <div className="moon-inner"></div>
            <div className="star star1"></div>
            <div className="star star2"></div>
            <div className="star star3"></div>
          </div>
        )}
      </div>
    </button>
  );
};

export default ToggleThemeButton;
