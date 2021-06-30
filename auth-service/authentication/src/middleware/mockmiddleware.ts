import { Request, Response, NextFunction } from 'express';
// mock middleware for testing purpose; not implemented in else where for now
export class MockMiddleWare {
    authenticateIbmId = async (req: Request, res: Response, next: NextFunction) => {
        req.user = {_json:{email : "test@sg.ibm.com"}};
        req.user.emailAddress = "test@sg.ibm.com";
        console.debug(req.user);
        next();
    }
}