import backupComplete from '../assets/images/character/backup-complete.png';
import dbError from '../assets/images/character/db-error.png';
import dbLoading from '../assets/images/character/db-loading.png';
import dbRunning from '../assets/images/character/db-running.png';
import highUsage from '../assets/images/character/high-usage.png';
import lowCredits from '../assets/images/character/low-credits.png';
import maintenance from '../assets/images/character/maintenance.png';
import richInCredits from '../assets/images/character/rich-credits.png';

export const characterImages = {
    backupComplete,
    error: dbError,
    loading: dbLoading,
    provisioning: dbLoading,
    running: dbRunning,
    stopped: dbError,
    highUsage,
    lowCredits,
    maintenance,
    richInCredits,
    default: dbRunning
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