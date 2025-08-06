import backupComplete from '../assets/images/character/backup-complete.png';
import dbError from '../assets/images/character/db-error.png';
import dbRunning from '../assets/images/character/db-running.png';
import emptyState from '../assets/images/character/empty-state.png'
import highUsage from '../assets/images/character/high-usage.png';
import lowCredits from '../assets/images/character/low-credits.png';
import maintenance from '../assets/images/character/maintenance.png';
import richInCredits from '../assets/images/character/rich-credits.png';
import React from "react";
import LoadingLemonSVG from "../components/common/LoadingCharacterSVG";

type CharacterImage = string | React.ComponentType<any>;

export const characterImages = {
    backupComplete,
    error: dbError,
    loading: LoadingLemonSVG,
    provisioning: LoadingLemonSVG,
    running: dbRunning,
    stopped: dbError,
    highUsage,
    lowCredits,
    maintenance,
    richInCredits,
    default: emptyState
};

export const getCharacterByStatus = (status: string) => {
    switch (status) {
        case 'running':
            return characterImages.running;
        case 'provisioning':
            return characterImages.provisioning;
        case 'error':
        case 'stopped':
            return characterImages.error;
        case 'maintenance':
            return characterImages.maintenance;
        default:
            return characterImages.default;
    }
};

export const isImageComponent = (image: CharacterImage): image is React.ComponentType<any> => {
    return typeof image === 'function';
};