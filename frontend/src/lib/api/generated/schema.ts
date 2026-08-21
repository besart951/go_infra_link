/* eslint-disable */
// This file is generated from backend/docs/swagger.json. Do not edit manually.

export type paths = {
    "/api/v1/account/notifications": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List current user's system notifications */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Only unread notifications */
                    unread_only?: boolean;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SystemNotificationListResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/account/notifications/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Delete one system notification */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Notification ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/account/notifications/{id}/important": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Toggle important state for one system notification */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Notification ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SystemNotificationResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/account/notifications/{id}/read": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Mark one system notification as read */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Notification ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SystemNotificationResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/account/notifications/{id}/read-toggle": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Toggle read state for one system notification */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Notification ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SystemNotificationResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/account/notifications/preferences": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get current user's notification preference */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.UserNotificationPreferenceResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        /** Create or update current user's notification preference */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Notification preference */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.UpsertUserNotificationPreferenceRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.UserNotificationPreferenceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/account/notifications/preferences/email-verification": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Send an email verification code for current user's notification email */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.UserNotificationPreferenceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/account/notifications/preferences/email-verification/verify": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Verify current user's notification email with a code */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Verification code */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.VerifyUserNotificationEmailRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.UserNotificationPreferenceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/account/notifications/read-all": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Mark all current user's system notifications as read */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/admin/notifications/smtp": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get SMTP notification settings */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SMTPSettingsResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        /** Create or update SMTP notification settings */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description SMTP settings */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.UpsertSMTPSettingsRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SMTPSettingsResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/admin/notifications/smtp/test": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Send an SMTP test email */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Test email payload */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SendSMTPTestEmailRequest"];
                };
            };
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/admin/users/{id}/disable": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Disable a user */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description User ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/admin/users/{id}/enable": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Enable a user */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description User ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/admin/users/{id}/restore": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Restore a deleted user */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description User ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/admin/users/{id}/role": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Set a user's role */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description User ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Role */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.AdminSetUserRoleRequest"];
                };
            };
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/auth/login": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Login */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Login data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.LoginRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.AuthResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/auth/logout": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Logout */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/auth/me": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get current user */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.AuthUserResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/auth/refresh": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Refresh access token */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.AuthResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/auth/session": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get current auth session status */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.SessionResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/alarm-definitions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List alarm definitions with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmDefinitionListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new alarm definition */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Alarm Definition data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateAlarmDefinitionRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmDefinitionResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/alarm-definitions/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get an alarm definition by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Definition ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmDefinitionResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update an alarm definition */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Definition ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Alarm Definition data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateAlarmDefinitionRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmDefinitionResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete an alarm definition */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Definition ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/alarm-fields": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List alarm fields */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmFieldListResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create an alarm field */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Alarm field data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateAlarmFieldRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmFieldResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/alarm-fields/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get an alarm field */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Field ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmFieldResponse"];
                    };
                };
            };
        };
        /** Update an alarm field */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Field ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Alarm field data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateAlarmFieldRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmFieldResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete an alarm field */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Field ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/alarm-type-fields/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** Update an alarm type field mapping */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Type Field ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Alarm type field data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateAlarmTypeFieldRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmTypeFieldResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete an alarm type field mapping */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Type Field ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/alarm-types": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List alarm types */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: never;
        };
        put?: never;
        /** Create an alarm type */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Alarm type data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateAlarmTypeRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmTypeResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/alarm-types/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get an alarm type */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Type ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmTypeResponse"];
                    };
                };
            };
        };
        /** Update an alarm type */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Type ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Alarm type data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateAlarmTypeRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmTypeResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete an alarm type */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Type ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/alarm-types/{id}/fields": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get fields for an alarm type */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Type ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmTypeResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Add a field to an alarm type */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Alarm Type ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Alarm type field data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateAlarmTypeFieldRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmTypeFieldResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/alarm-units": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List alarm units */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UnitListResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create an alarm unit */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Unit data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateUnitRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UnitResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/alarm-units/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get an alarm unit */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Unit ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UnitResponse"];
                    };
                };
            };
        };
        /** Update an alarm unit */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Unit ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Unit data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateUnitRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UnitResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete an alarm unit */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Unit ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/apparats": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List apparats with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Filter by Object Data ID */
                    object_data_id?: string;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                    /** @description Filter by System Part ID */
                    system_part_id?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new apparat */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Apparat data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateApparatRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/apparats/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get an apparat by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Apparat ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update an apparat */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Apparat ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Apparat data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateApparatRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete an apparat */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Apparat ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/apparats/bulk": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Get multiple apparats by IDs */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Apparat IDs */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatBulkRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatBulkResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/bacnet-objects": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Create a bacnet object (for field device or object data) */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Bacnet Object data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateBacnetObjectRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/bacnet-objects/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a BACnet instance or template by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description BACnet Object ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update a bacnet object */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Bacnet Object ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Bacnet Object data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateBacnetObjectRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/bacnet-objects/{id}/alarm-schema": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get alarm field schema for a BacnetObject */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description BacnetObject ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmTypeResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/bacnet-objects/{id}/alarm-values": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get alarm values for a BacnetObject */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description BacnetObject ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmValuesResponse"];
                    };
                };
            };
        };
        /** Replace alarm values for a BacnetObject */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description BacnetObject ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Alarm values */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.PutAlarmValuesRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmValuesResponse"];
                    };
                };
            };
        };
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/bacnet-reference-usages": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Count BACnet object usage for reference data */
        get: {
            parameters: {
                query: {
                    /** @description Reference IDs */
                    ids: string[];
                    /** @description Reference resource */
                    resource: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetReferenceUsageListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/buildings": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List buildings with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new building */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Building data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateBuildingRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/buildings/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a building by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Building ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update a building */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Building ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateBuildingRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a building */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Building ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        /** Update a building */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Building ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateBuildingRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/facility/buildings/{id}/detail": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a building detail with permitted hierarchy relations */
        get: {
            parameters: {
                query?: {
                    /** @description Relationship page size */
                    limit?: number;
                    /** @description Relationship page */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description Building ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingDetailResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/buildings/bulk": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Get multiple buildings by IDs */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Building IDs */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingBulkRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingBulkResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/buildings/validate": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Validate building fields */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Validation payload */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ValidateBuildingRequest"];
                };
            };
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/control-cabinets": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List control cabinets with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Building ID */
                    building_id?: string;
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new control cabinet */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Control Cabinet data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateControlCabinetRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/control-cabinets/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a control cabinet by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Control Cabinet ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update a control cabinet */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Control Cabinet ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateControlCabinetRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a control cabinet */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Control Cabinet ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        /** Update a control cabinet */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Control Cabinet ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateControlCabinetRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/facility/control-cabinets/{id}/copy": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Copy a control cabinet */
        post: {
            parameters: {
                query?: never;
                header?: {
                    /** @description Client-generated copy operation UUID */
                    "X-Copy-Operation-ID"?: string;
                };
                path: {
                    /** @description Control Cabinet ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Accepted */
                202: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CopyJobResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Service Unavailable */
                503: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/control-cabinets/{id}/delete-impact": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Preview delete impact for a control cabinet */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Control Cabinet ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetDeleteImpactResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/control-cabinets/{id}/detail": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a control cabinet detail with permitted hierarchy relations */
        get: {
            parameters: {
                query?: {
                    /** @description Relationship page size */
                    limit?: number;
                    /** @description Relationship page */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description Control cabinet ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetDetailResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/control-cabinets/bulk": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Get multiple control cabinets by IDs */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Control Cabinet IDs */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetBulkRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetBulkResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/control-cabinets/validate": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Validate control cabinet fields */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Validation payload */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ValidateControlCabinetRequest"];
                };
            };
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/copy-jobs/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a facility copy job */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Copy Job ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CopyJobResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/delete-impacts": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Preview blocking facility references before deletion */
        get: {
            parameters: {
                query: {
                    /** @description Reference IDs */
                    ids: string[];
                    /** @description Reference resource (apparat or system_part) */
                    resource: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DeleteImpactListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/field-devices": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List field devices with pagination and filtering */
        get: {
            parameters: {
                query?: {
                    /** @description Filter by building ID(s), accepts a single UUID or a | separated list */
                    building_id?: string;
                    /** @description Filter by control cabinet ID(s), accepts a single UUID or a | separated list */
                    control_cabinet_id?: string;
                    /** @description Items per page */
                    limit?: number;
                    /** @description Order direction (asc, desc) */
                    order?: string;
                    /** @description Order by (created_at,sps_system_type,bmk,description,apparat_nr,apparat,system_part,spec_supplier,spec_brand,spec_type,spec_motor_valve,spec_size,spec_install_loc,spec_ph,spec_acdc,spec_amperage,spec_power,spec_rotation) */
                    order_by?: string;
                    /** @description Page number */
                    page?: number;
                    /** @description Filter by project ID */
                    project_id?: string;
                    /** @description Search query */
                    search?: string;
                    /** @description Filter by SPS controller ID(s), accepts a single UUID or a | separated list */
                    sps_controller_id?: string;
                    /** @description Filter by SPS controller system type ID(s), accepts a single UUID or a | separated list */
                    sps_controller_system_type_id?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/field-devices/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a field device by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field Device ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update a field device */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field Device ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateFieldDeviceRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a field device */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field Device ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        /** Update a field device */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field Device ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateFieldDeviceRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/facility/field-devices/{id}/bacnet-objects": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List bacnet objects for a field device (hydration) */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field Device ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectResponse"][];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/field-devices/{id}/copy": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Deep-copy a field device */
        post: {
            parameters: {
                query?: never;
                header?: {
                    /** @description Client-generated operation UUID */
                    "Idempotency-Key"?: string;
                };
                path: {
                    /** @description Field Device ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Accepted */
                202: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FacilityJobResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Service Unavailable */
                503: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/field-devices/{id}/detail": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a field device detail with permitted hierarchy and references */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field device ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceDetailResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/field-devices/{id}/specification": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get the specification owned by a field device */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field Device ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SpecificationResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update specification for a field device */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field Device ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Specification data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateFieldDeviceSpecificationRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SpecificationResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Create specification for a field device */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field Device ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Specification data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateFieldDeviceSpecificationRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SpecificationResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Delete the specification owned by a field device */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field Device ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/field-devices/available-apparat-nr": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List available apparat numbers for field devices */
        get: {
            parameters: {
                query: {
                    /** @description Apparat ID */
                    apparat_id: string;
                    /** @description SPS Controller System Type ID */
                    sps_controller_system_type_id: string;
                    /** @description System Part ID */
                    system_part_id: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AvailableApparatNumbersResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/field-devices/bulk-delete": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Bulk delete multiple field devices */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Bulk delete request */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkDeleteFieldDeviceRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkDeleteFieldDeviceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/field-devices/bulk-update": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /**
         * Bulk update multiple field devices
         * @description Updates multiple field devices in a single operation. Supports nested specification and BACnet objects updates.
         */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Bulk update request */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkUpdateFieldDeviceRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkUpdateFieldDeviceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/facility/field-devices/multi-create": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Create multiple field devices in a single operation
         * @description Creates multiple field devices with independent validation. Returns detailed results for each device. To link created devices to a project, use the CreateProjectFieldDevice endpoint with the returned field device IDs.
         */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Multi-create request */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.MultiCreateFieldDeviceRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.MultiCreateFieldDeviceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/field-devices/options": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Get all metadata needed for creating/editing field devices
         * @description Returns all apparats, system parts, object datas and their relationships in a single call. This returns global templates (object data where project_id is null and is_active = true).
         */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceOptionsResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/jobs": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List the current user's facility jobs */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FacilityJobListResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/notification-classes": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List notification classes with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.NotificationClassListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new notification class */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Notification Class data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateNotificationClassRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.NotificationClassResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/notification-classes/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a notification class by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Notification Class ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.NotificationClassResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update a notification class */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Notification Class ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Notification Class data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateNotificationClassRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.NotificationClassResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a notification class */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Notification Class ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/object-data": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List object data with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Filter by Apparat ID */
                    apparat_id?: string;
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                    /** @description Filter by System Part ID */
                    system_part_id?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ObjectDataListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create object data */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Object Data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateObjectDataRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ObjectDataResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/object-data/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get object data by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Object Data ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ObjectDataResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update object data */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Object Data ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Object Data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateObjectDataRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ObjectDataResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete object data */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Object Data ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/object-data/{id}/bacnet-objects": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get bacnet objects for object data */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Object Data ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectResponse"][];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/object-data/{id}/copy": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Deep-copy an object data template */
        post: {
            parameters: {
                query?: never;
                header?: {
                    /** @description Client-generated operation UUID */
                    "Idempotency-Key"?: string;
                };
                path: {
                    /** @description Object Data ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Accepted */
                202: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FacilityJobResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Service Unavailable */
                503: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/reference-data/stream": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Stream facility reference-data changes
         * @description Upgrades the authenticated request to the shared facility WebSocket. `facility_reference_data.changed` tells authorized clients to refresh cached apparats and system parts. `facility.changed` carries authorized facility resource changes with an action, IDs, actor and timestamp. User-scoped `facility.copy_job.progress` events contain a copy job ID, status, stage and 0-100 progress; they are only delivered to the user that started the job.
         */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Switching Protocols */
                101: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/sps-controller-system-types": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List SPS controller system types with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Project ID (filter by project) */
                    project_id?: string;
                    /** @description Search query */
                    search?: string;
                    /** @description SPS Controller ID(s), accepts a single UUID or a | separated list */
                    sps_controller_id?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.SPSControllerSystemTypeListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create an SPS controller system type assignment */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description SPS controller system type data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateSPSControllerSystemTypeRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.SPSControllerSystemTypeResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/sps-controller-system-types/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get an SPS controller system type by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description SPS Controller System Type ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.SPSControllerSystemTypeResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update an SPS controller system type */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description SPS Controller System Type ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerSystemTypeRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.SPSControllerSystemTypeResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete an SPS controller system type */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description SPS Controller System Type ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        /** Update an SPS controller system type */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description SPS Controller System Type ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerSystemTypeRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.SPSControllerSystemTypeResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/facility/sps-controller-system-types/{id}/copy": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Copy an SPS controller system type */
        post: {
            parameters: {
                query?: never;
                header?: {
                    /** @description Client-generated copy operation UUID */
                    "X-Copy-Operation-ID"?: string;
                };
                path: {
                    /** @description SPS Controller System Type ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Accepted */
                202: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CopyJobResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
                /** @description Service Unavailable */
                503: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/sps-controller-system-types/{id}/detail": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get an SPS controller system type detail with permitted hierarchy relations */
        get: {
            parameters: {
                query?: {
                    /** @description Relationship page size */
                    limit?: number;
                    /** @description Relationship page */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description SPS controller system type ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeDetailResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/sps-controllers": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List SPS controllers with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Control Cabinet ID */
                    control_cabinet_id?: string;
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new SPS controller */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description SPS Controller data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateSPSControllerRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/sps-controllers/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get an SPS controller by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description SPS Controller ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update an SPS controller */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description SPS Controller ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete an SPS controller */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description SPS Controller ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        /** Update an SPS controller */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description SPS Controller ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/facility/sps-controllers/{id}/copy": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Copy an SPS controller */
        post: {
            parameters: {
                query?: never;
                header?: {
                    /** @description Client-generated copy operation UUID */
                    "X-Copy-Operation-ID"?: string;
                };
                path: {
                    /** @description SPS Controller ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Accepted */
                202: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CopyJobResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Service Unavailable */
                503: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/sps-controllers/{id}/detail": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get an SPS controller detail with permitted hierarchy relations */
        get: {
            parameters: {
                query?: {
                    /** @description Relationship page size */
                    limit?: number;
                    /** @description Relationship page */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description SPS controller ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerDetailResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/sps-controllers/bulk": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Get multiple SPS controllers by IDs */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description SPS Controller IDs */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerBulkRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerBulkResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/sps-controllers/next-ga-device": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Suggest next available GA device for a control cabinet */
        get: {
            parameters: {
                query: {
                    /** @description Control Cabinet ID */
                    control_cabinet_id: string;
                    /** @description SPS Controller ID to exclude */
                    exclude_id?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.NextAvailableGADeviceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/sps-controllers/validate": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Validate SPS controller fields */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Validation payload */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ValidateSPSControllerRequest"];
                };
            };
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/state-texts": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List state texts with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.StateTextListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new state text */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description State Text data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateStateTextRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.StateTextResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/state-texts/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a state text by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description State Text ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.StateTextResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update a state text */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description State Text ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description State Text data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateStateTextRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.StateTextResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a state text */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description State Text ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/system-parts": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List system parts with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Filter by Apparat ID */
                    apparat_id?: string;
                    /** @description Items per page */
                    limit?: number;
                    /** @description Filter by Object Data ID */
                    object_data_id?: string;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemPartListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new system part */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description System Part data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateSystemPartRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemPartResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/system-parts/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a system part by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description System Part ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemPartResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /** Update a system part */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description System Part ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description System Part data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSystemPartRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemPartResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a system part */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description System Part ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/system-types": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List system types with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemTypeListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /**
         * Create a new system type
         * @description number_min and number_max must not overlap existing ranges. number_min may equal number_max.
         */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description System Type data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateSystemTypeRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemTypeResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/facility/system-types/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a system type by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description System Type ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemTypeResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        /**
         * Update a system type
         * @description number_min and number_max must not overlap existing ranges. number_min may equal number_max.
         */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description System Type ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description System Type data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSystemTypeRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemTypeResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a system type */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description System Type ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/history/timeline": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * List global audit activities
         * @description Returns authoritative audit events with their actual before/after diff. Multiple action and field parameters are combined as OR filters within their category.
         */
        get: {
            parameters: {
                query?: {
                    /** @description Actions: create, update, delete, restore */
                    action?: string[];
                    /** @description Actor UUID */
                    actor_id?: string;
                    /** @description Entity UUID */
                    entity_id?: string;
                    /** @description Entity table */
                    entity_table?: string;
                    /** @description Changed field names */
                    field?: string[];
                    /** @description Items per page */
                    limit?: number;
                    /** @description Earliest ISO-8601 timestamp */
                    occurred_from?: string;
                    /** @description Latest ISO-8601 timestamp */
                    occurred_to?: string;
                    /** @description Page number */
                    page?: number;
                    /** @description Scope UUID */
                    scope_id?: string;
                    /** @description Scope type */
                    scope_type?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.TimelineResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ErrorResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/i18n/{locale}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get translations for a specific locale */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Locale code (e.g., de_CH, en_US) */
                    locale: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Translation data */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            [key: string]: unknown;
                        };
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_i18n.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_i18n.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["internal_handler_i18n.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/permissions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List permission types */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.PermissionResponse"][];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a permission type */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Permission data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.CreatePermissionRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.PermissionResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/permissions/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** Update a permission type */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Permission ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Permission data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UpdatePermissionRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.PermissionResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a permission type */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Permission ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/phases": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List phases with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.PhaseListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new phase */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Phase data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreatePhaseRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.PhaseResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/phases/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a phase by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Phase ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.PhaseResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        /**
         * Update a phase
         * @description PATCH-like update: omitted fields remain unchanged and present string fields are applied even when empty. PUT is kept for compatibility; PATCH is the preferred method.
         */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Phase ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdatePhaseRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.PhaseResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a phase */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Phase ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        /**
         * Update a phase
         * @description PATCH-like update: omitted fields remain unchanged and present string fields are applied even when empty. PUT is kept for compatibility; PATCH is the preferred method.
         */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Phase ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdatePhaseRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.PhaseResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/projects": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List projects with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Phase ID filter */
                    phase_id?: string;
                    /** @description Search query */
                    search?: string;
                    /** @description Status filter */
                    status?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new project */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Project data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a project by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        /**
         * Update a project
         * @description PATCH-like update: omitted fields remain unchanged, present string fields are applied even when empty, and start_date:null clears the date. PUT is kept for compatibility; PATCH is the preferred method.
         */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a project */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        /**
         * Update a project
         * @description PATCH-like update: omitted fields remain unchanged, present string fields are applied even when empty, and start_date:null clears the date. PUT is kept for compatibility; PATCH is the preferred method.
         */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody: components["requestBodies"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectRequest"];
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/projects/{id}/capabilities": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Get effective permissions for a project
         * @description Returns project-scoped permissions after project membership and phase restrictions have been applied.
         */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectCapabilitiesResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/changes": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * List project changes after a revision
         * @description Returns durable project changes for HTTP recovery after a missed collaboration event.
         */
        get: {
            parameters: {
                query?: {
                    /** @description Last processed revision */
                    after_revision?: number;
                    /** @description Maximum number of changes */
                    limit?: number;
                };
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectChangesResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/collaboration": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Stream project collaboration events
         * @description Upgrades the authenticated request to a WebSocket. Messages use the frontend's ProjectSyncInboundMessage contract.
         */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Switching Protocols */
                101: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["internal_handler_project.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["internal_handler_project.ErrorResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["internal_handler_project.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/control-cabinets": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List project control cabinets with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectControlCabinetListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create project control cabinet link */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Link data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectControlCabinetRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectControlCabinetResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/control-cabinets/{controlCabinetId}/copy": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Copy a project control cabinet asynchronously */
        post: {
            parameters: {
                query?: never;
                header?: {
                    /** @description Client-generated copy operation UUID */
                    "X-Copy-Operation-ID"?: string;
                };
                path: {
                    /** @description Control Cabinet ID */
                    controlCabinetId: string;
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Accepted */
                202: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CopyJobResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Service Unavailable */
                503: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/control-cabinets/{linkId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** Update project control cabinet link */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description Link ID */
                    linkId: string;
                };
                cookie?: never;
            };
            /** @description Link data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectControlCabinetRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectControlCabinetResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete project control cabinet link */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description Link ID */
                    linkId: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/facility/buildings/{buildingId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a project-scoped, read-only building detail */
        get: {
            parameters: {
                query?: {
                    /** @description Relationship page size */
                    limit?: number;
                    /** @description Relationship page */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description Building ID */
                    buildingId: string;
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingDetailResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/facility/control-cabinets/{controlCabinetId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a project-scoped control cabinet detail */
        get: {
            parameters: {
                query?: {
                    /** @description Relationship page size */
                    limit?: number;
                    /** @description Relationship page */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description Control cabinet ID */
                    controlCabinetId: string;
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetDetailResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** Patch a project control cabinet when both permission layers allow it */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Control cabinet ID */
                    controlCabinetId: string;
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Control cabinet patch */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateControlCabinetRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/projects/{id}/facility/field-devices/{fieldDeviceId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a project-scoped field device detail */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field device ID */
                    fieldDeviceId: string;
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceDetailResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** Patch a project field device when both permission layers allow it */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Field device ID */
                    fieldDeviceId: string;
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Field device patch */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateFieldDeviceRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/projects/{id}/facility/sps-controller-system-types/{spsControllerSystemTypeId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a project-scoped SPS controller system type detail */
        get: {
            parameters: {
                query?: {
                    /** @description Relationship page size */
                    limit?: number;
                    /** @description Relationship page */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description SPS controller system type ID */
                    spsControllerSystemTypeId: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeDetailResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** Patch a project SPS controller system type when both permission layers allow it */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description SPS controller system type ID */
                    spsControllerSystemTypeId: string;
                };
                cookie?: never;
            };
            /** @description System type patch */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerSystemTypeRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/projects/{id}/facility/sps-controllers/{spsControllerId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a project-scoped SPS controller detail */
        get: {
            parameters: {
                query?: {
                    /** @description Relationship page size */
                    limit?: number;
                    /** @description Relationship page */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description SPS controller ID */
                    spsControllerId: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerDetailResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** Patch a project SPS controller when both permission layers allow it */
        patch: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description SPS controller ID */
                    spsControllerId: string;
                };
                cookie?: never;
            };
            /** @description SPS controller patch */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse"];
                    };
                };
            };
        };
        trace?: never;
    };
    "/api/v1/projects/{id}/field-device-options": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Get all metadata needed for creating/editing field devices within a project
         * @description Returns all apparats, system parts, object datas and their relationships for a specific project. This returns project-specific object data (object data where project_id = :id and is_active = true).
         */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceOptionsResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/field-devices": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List project field devices with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectFieldDeviceListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create project field device link */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Link data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectFieldDeviceRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectFieldDeviceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/field-devices/{linkId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** Update project field device link */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description Link ID */
                    linkId: string;
                };
                cookie?: never;
            };
            /** @description Link data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectFieldDeviceRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectFieldDeviceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete project field device link */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description Link ID */
                    linkId: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/field-devices/multi-create": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Create multiple project field device links */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Link data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.MultiCreateProjectFieldDeviceRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.MultiCreateProjectFieldDeviceResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/history/timeline": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * List project audit activities
         * @description Returns authoritative audit events for one project. The optional entity, field, action, actor and date filters work like the global timeline.
         */
        get: {
            parameters: {
                query?: {
                    /** @description Actions: create, update, delete, restore */
                    action?: string[];
                    /** @description Actor UUID */
                    actor_id?: string;
                    /** @description Entity UUID */
                    entity_id?: string;
                    /** @description Entity table */
                    entity_table?: string;
                    /** @description Changed field names */
                    field?: string[];
                    /** @description Items per page */
                    limit?: number;
                    /** @description Earliest ISO-8601 timestamp */
                    occurred_from?: string;
                    /** @description Latest ISO-8601 timestamp */
                    occurred_to?: string;
                    /** @description Page number */
                    page?: number;
                    /** @description Secondary scope UUID */
                    scope_id?: string;
                    /** @description Secondary scope type */
                    scope_type?: string;
                };
                header?: never;
                path: {
                    /** @description Project UUID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.TimelineResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ErrorResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/object-data": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List project object data with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Filter by Apparat ID */
                    apparat_id?: string;
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                    /** @description Filter by System Part ID */
                    system_part_id?: string;
                };
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ObjectDataListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Attach object data to project */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Object data link */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectObjectDataRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ObjectDataResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Conflict */
                409: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/object-data/{objectDataId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Detach object data from project */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description Object Data ID */
                    objectDataId: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ObjectDataResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/sps-controller-system-types/{systemTypeId}/copy": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Copy a project SPS controller system type asynchronously */
        post: {
            parameters: {
                query?: never;
                header?: {
                    /** @description Client-generated copy operation UUID */
                    "X-Copy-Operation-ID"?: string;
                };
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description SPS Controller System Type ID */
                    systemTypeId: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Accepted */
                202: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CopyJobResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Service Unavailable */
                503: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/sps-controllers": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List project SPS controllers with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectSPSControllerListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create project SPS controller link */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Link data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectSPSControllerRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectSPSControllerResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/sps-controllers/{linkId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** Update project SPS controller link */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description Link ID */
                    linkId: string;
                };
                cookie?: never;
            };
            /** @description Link data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectSPSControllerRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectSPSControllerResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete project SPS controller link */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description Link ID */
                    linkId: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/sps-controllers/{spsControllerId}/copy": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Copy a project SPS controller asynchronously */
        post: {
            parameters: {
                query?: never;
                header?: {
                    /** @description Client-generated copy operation UUID */
                    "X-Copy-Operation-ID"?: string;
                };
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description SPS Controller ID */
                    spsControllerId: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Accepted */
                202: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CopyJobResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Service Unavailable */
                503: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/users": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List users in a project */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectUserListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Invite user to project */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Invite data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectUserRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectUserResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/{id}/users/{userId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Remove user from project */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Project ID */
                    id: string;
                    /** @description User ID */
                    userId: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/roles": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List roles with permissions */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.RoleResponse"][];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/roles/{role}/permissions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** Replace permissions for a role */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Role */
                    role: string;
                };
                cookie?: never;
            };
            /** @description Role permissions */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UpdateRolePermissionsRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.RoleResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        /** Assign a permission to a role */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Role */
                    role: string;
                };
                cookie?: never;
            };
            /** @description Permission */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.AddRolePermissionRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.RolePermissionResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/roles/{role}/permissions/{permission}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Remove a permission from a role */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Permission name */
                    permission: string;
                    /** @description Role */
                    role: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/teams": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List teams with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.TeamListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new team */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Team data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.CreateTeamRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.TeamResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/teams/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a team by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Team ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.TeamResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
            };
        };
        /** Update a team */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Team ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Team data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.UpdateTeamRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.TeamResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a team */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Team ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/teams/{id}/members": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List team members */
        get: {
            parameters: {
                query?: {
                    /** @description Items per page */
                    limit?: number;
                    /** @description Page number */
                    page?: number;
                };
                header?: never;
                path: {
                    /** @description Team ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.TeamMemberListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Add a member to a team */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Team ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description Member data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.AddTeamMemberRequest"];
                };
            };
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/teams/{id}/members/{userId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Remove a member from a team */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description Team ID */
                    id: string;
                    /** @description User ID */
                    userId: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "*/*": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/users": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List users with pagination */
        get: {
            parameters: {
                query?: {
                    /** @description Include soft-deleted users */
                    include_deleted?: boolean;
                    /** @description Items per page */
                    limit?: number;
                    /** @description Sort direction */
                    order?: "asc" | "desc";
                    /** @description Sort field */
                    order_by?: string;
                    /** @description Page number */
                    page?: number;
                    /** @description Search query */
                    search?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserListResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        /** Create a new user */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description User data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.CreateUserRequest"];
                };
            };
            responses: {
                /** @description Created */
                201: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/users/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a user by ID */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description User ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        /** Update a user */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description User ID */
                    id: string;
                };
                cookie?: never;
            };
            /** @description User data */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UpdateUserRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Not Found */
                404: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        /** Delete a user */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    /** @description User ID */
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description No Content */
                204: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/users/allowed-roles": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get roles that the current user can assign */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.AllowedRolesResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/users/directory": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List visible users for the user directory */
        get: {
            parameters: {
                query?: {
                    /** @description Include soft-deleted users */
                    include_deleted?: boolean;
                    /** @description Items per page */
                    limit?: number;
                    /** @description Sort direction */
                    order?: "asc" | "desc";
                    /** @description Sort field */
                    order_by?: string;
                    /** @description Page number */
                    page?: number;
                    /** @description Role filter */
                    role?: string;
                    /** @description Search query */
                    search?: string;
                    /** @description Visible team filter */
                    team_id?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryListResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Forbidden */
                403: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/users/me/password": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** Update own password */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            /** @description Password */
            requestBody: {
                content: {
                    "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UpdateOwnPasswordRequest"];
                };
            };
            responses: {
                /** @description OK */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserResponse"];
                    };
                };
                /** @description Bad Request */
                400: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Unauthorized */
                401: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
                /** @description Internal Server Error */
                500: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse"];
                    };
                };
            };
        };
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
};
export type webhooks = Record<string, never>;
export type components = {
    schemas: {
        /** @enum {string} */
        "github_com_besart951_go_infra_link_backend_internal_domain_project.ProjectStatus": "planned" | "ongoing" | "completed";
        /** @enum {string} */
        "github_com_besart951_go_infra_link_backend_internal_domain_user.Role": "superadmin" | "admin_fzag" | "fzag" | "admin_planer" | "planer" | "admin_entrepreneur" | "entrepreneur";
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.AuthResponse": {
            access_token_expires_at?: string;
            csrf_token?: string;
            refresh_token_expires_at?: string;
            user?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.AuthUserResponse"];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.AuthUserResponse": {
            can_access_user_directory?: boolean;
            created_at?: string;
            disabled_at?: string;
            email?: string;
            failed_login_attempts?: number;
            first_name?: string;
            id?: string;
            is_active?: boolean;
            last_login_at?: string;
            last_name?: string;
            locked_until?: string;
            permissions?: string[];
            role?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.ErrorResponse": {
            code?: string;
            conflict?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.WriteConflictResponse"];
            details?: unknown;
            /** @description Error and Fields are kept as compatibility aliases for existing clients. */
            error?: string;
            field_errors?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.FieldErrorResponse"][];
            fields?: {
                [key: string]: string;
            };
            localized_key?: string;
            message?: string;
            request_id?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.LoginRequest": {
            email: string;
            password: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_auth.SessionResponse": {
            authenticated?: boolean;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_common.FieldErrorResponse": {
            code?: string;
            localized_key?: string;
            message?: string;
            params?: {
                [key: string]: unknown;
            };
            path?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_common.WriteConflictResponse": {
            aggregate_id?: string;
            aggregate_type?: string;
            base_version?: number;
            conflicting_fields?: string[];
            current?: unknown;
            current_version?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmDefinitionListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmDefinitionResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmDefinitionResponse": {
            alarm_note?: string;
            alarm_type_id?: string;
            created_at?: string;
            id?: string;
            name?: string;
            scope?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmFieldListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmFieldResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmFieldResponse": {
            data_type?: string;
            default_unit_code?: string;
            id?: string;
            key?: string;
            label?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmTypeFieldResponse": {
            alarm_field?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmFieldResponse"];
            alarm_field_id?: string;
            alarm_type_id?: string;
            created_at?: string;
            default_unit?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UnitResponse"];
            default_unit_id?: string;
            default_value_json?: string;
            display_order?: number;
            id?: string;
            is_required?: boolean;
            is_user_editable?: boolean;
            ui_group?: string;
            updated_at?: string;
            validation_json?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmTypeResponse": {
            code?: string;
            created_at?: string;
            fields?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmTypeFieldResponse"][];
            id?: string;
            name?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmValueInput": {
            alarm_type_field_id: string;
            source?: string;
            unit_id?: string;
            value_boolean?: boolean;
            value_integer?: number;
            value_json?: string;
            value_number?: number;
            value_string?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmValueResponse": {
            alarm_type_field_id?: string;
            bacnet_object_id?: string;
            created_at?: string;
            id?: string;
            source?: string;
            unit_id?: string;
            updated_at?: string;
            value_boolean?: boolean;
            value_integer?: number;
            value_json?: string;
            value_number?: number;
            value_string?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmValuesResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmValueResponse"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatBulkRequest": {
            ids: string[];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatBulkResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatResponse"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatResponse": {
            created_at?: string;
            description?: string;
            id?: string;
            name?: string;
            short_name?: string;
            system_parts?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemPartResponse"][];
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AvailableApparatNumbersResponse": {
            available?: number[];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectBulkPatchInput": {
            alarm_definition_id?: string;
            alarm_type_id?: string;
            description?: string;
            gms_visible?: boolean;
            hardware_quantity?: number;
            hardware_type?: string;
            id: string;
            notification_class_id?: string;
            optional?: boolean;
            software_number?: number;
            software_reference_id?: string;
            software_type?: string;
            state_text_id?: string;
            text_fix?: string;
            text_individual?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectInput": {
            alarm_definition_id?: string;
            alarm_type_id?: string;
            description?: string;
            gms_visible?: boolean;
            hardware_quantity?: number;
            hardware_type?: string;
            notification_class_id?: string;
            optional?: boolean;
            software_number: number;
            software_reference_id?: string;
            software_type: string;
            state_text_id?: string;
            text_fix: string;
            text_individual?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectResponse": {
            /** @description AggregateVersion is the owning FieldDevice concurrency token. */
            aggregate_version?: number;
            alarm_definition_id?: string;
            alarm_type_id?: string;
            created_at?: string;
            description?: string;
            field_device_id?: string;
            gms_visible?: boolean;
            hardware_quantity?: number;
            hardware_type?: string;
            id?: string;
            notification_class_id?: string;
            optional?: boolean;
            software_number?: number;
            software_reference_id?: string;
            software_type?: string;
            state_text_id?: string;
            text_fix?: string;
            text_individual?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetReferenceUsageListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetReferenceUsageResponse"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetReferenceUsageResponse": {
            bacnet_object_count?: number;
            id?: string;
            resource?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingBulkRequest": {
            ids: string[];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingBulkResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingResponse"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingDetailResponse": {
            building?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingResponse"];
            capabilities?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailCapabilities"];
            relations?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailRelation"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingResponse": {
            building_group?: number;
            created_at?: string;
            id?: string;
            iws_code?: string;
            updated_at?: string;
            version?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkDeleteFieldDeviceRequest": {
            ids: string[];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkDeleteFieldDeviceResponse": {
            failure_count?: number;
            results?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkOperationResultItem"][];
            success_count?: number;
            total_count?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkOperationResultItem": {
            error?: string;
            field_device?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceResponse"];
            fields?: {
                [key: string]: string;
            };
            id?: string;
            merged?: boolean;
            success?: boolean;
            suggestion_options?: {
                [key: string]: number[];
            };
            suggestions?: {
                [key: string]: number;
            };
            version?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkUpdateFieldDeviceItem": {
            apparat_id?: string;
            apparat_nr?: number;
            bacnet_objects?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectBulkPatchInput"][];
            base_version?: number;
            bmk?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            description?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            id: string;
            specification?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SpecificationInput"];
            system_part_id?: string;
            text_fix?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkUpdateFieldDeviceRequest": {
            updates: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkUpdateFieldDeviceItem"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkUpdateFieldDeviceResponse": {
            failure_count?: number;
            results?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BulkOperationResultItem"][];
            success_count?: number;
            total_count?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetBulkRequest": {
            ids: string[];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetBulkResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetResponse"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetDeleteImpactResponse": {
            bacnet_objects_count?: number;
            control_cabinet_id?: string;
            field_devices_count?: number;
            specifications_count?: number;
            sps_controller_system_types_count?: number;
            sps_controllers_count?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetDetailResponse": {
            capabilities?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailCapabilities"];
            control_cabinet?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetResponse"];
            relations?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailRelation"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetResponse": {
            building_id?: string;
            control_cabinet_nr?: string;
            created_at?: string;
            id?: string;
            updated_at?: string;
            version?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CopyJobResponse": {
            attempts?: number;
            /** @enum {string} */
            class?: "mutation" | "export";
            completed_at?: string;
            created_at?: string;
            error?: string;
            failure_count?: number;
            job_id?: string;
            /** @enum {string} */
            kind?: "control_cabinet" | "sps_controller" | "sps_controller_system_type" | "field_device" | "object_data";
            processed?: number;
            progress?: number;
            result?: Record<string, never>;
            retryable?: boolean;
            stage?: string;
            /** @enum {string} */
            status?: "queued" | "running" | "completed" | "failed";
            success_count?: number;
            total?: number;
            /** @enum {string} */
            type?: "copy" | "export" | "bulk" | "delete" | "restore";
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateAlarmDefinitionRequest": {
            alarm_note?: string;
            alarm_type_id?: string;
            name: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateAlarmFieldRequest": {
            data_type: string;
            default_unit_code?: string;
            key: string;
            label: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateAlarmTypeFieldRequest": {
            alarm_field_id: string;
            default_unit_id?: string;
            default_value_json?: string;
            display_order?: number;
            is_required?: boolean;
            is_user_editable?: boolean;
            ui_group?: string;
            validation_json?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateAlarmTypeRequest": {
            code: string;
            name: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateApparatRequest": {
            description?: string;
            name: string;
            short_name: string;
            system_part_ids?: string[];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateBacnetObjectRequest": {
            alarm_definition_id?: string;
            alarm_type_id?: string;
            description?: string;
            field_device_id?: string;
            gms_visible?: boolean;
            hardware_quantity?: number;
            hardware_type?: string;
            notification_class_id?: string;
            object_data_id?: string;
            optional?: boolean;
            software_number: number;
            software_reference_id?: string;
            software_type: string;
            state_text_id?: string;
            text_fix: string;
            text_individual?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateBuildingRequest": {
            building_group: number;
            iws_code: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateControlCabinetRequest": {
            building_id: string;
            control_cabinet_nr: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateFieldDeviceRequest": {
            apparat_id: string;
            apparat_nr: number;
            bacnet_objects?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectInput"][];
            bmk?: string;
            description?: string;
            object_data_id?: string;
            sps_controller_system_type_id: string;
            system_part_id: string;
            text_fix?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateFieldDeviceSpecificationRequest": {
            additional_info_motor_valve?: string;
            additional_info_size?: number;
            additional_information_installation_location?: string;
            electrical_connection_acdc?: string;
            electrical_connection_amperage?: number;
            electrical_connection_ph?: number;
            electrical_connection_power?: number;
            electrical_connection_rotation?: number;
            specification_brand?: string;
            specification_supplier?: string;
            specification_type?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateNotificationClassRequest": {
            ack_required_error?: boolean;
            ack_required_normal?: boolean;
            ack_required_not_normal?: boolean;
            event_category: string;
            internal_description: string;
            meaning: string;
            nc: number;
            norm_error?: number;
            norm_normal?: number;
            norm_not_normal?: number;
            object_description: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateObjectDataRequest": {
            apparat_ids?: string[];
            bacnet_objects?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectInput"][];
            description: string;
            is_active?: boolean;
            project_id?: string;
            version: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateSPSControllerRequest": {
            control_cabinet_id: string;
            device_description?: string;
            device_location?: string;
            device_name: string;
            ga_device: string;
            gateway?: string;
            ip_address?: string;
            subnet?: string;
            system_types?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeInput"][];
            vlan?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateSPSControllerSystemTypeRequest": {
            document_name?: string;
            number?: number;
            sps_controller_id: string;
            system_type_id: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateStateTextRequest": {
            ref_number: number;
            state_text1?: string;
            state_text2?: string;
            state_text3?: string;
            state_text4?: string;
            state_text5?: string;
            state_text6?: string;
            state_text7?: string;
            state_text8?: string;
            state_text9?: string;
            state_text10?: string;
            state_text11?: string;
            state_text12?: string;
            state_text13?: string;
            state_text14?: string;
            state_text15?: string;
            state_text16?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateSystemPartRequest": {
            description?: string;
            name: string;
            short_name: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateSystemTypeRequest": {
            name: string;
            number_max: number;
            number_min: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateUnitRequest": {
            code: string;
            name: string;
            symbol: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DeleteImpactBlockerResponse": {
            count?: number;
            resource?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DeleteImpactListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DeleteImpactResponse"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DeleteImpactResponse": {
            blockers?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DeleteImpactBlockerResponse"][];
            id?: string;
            resource?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailCapabilities": {
            can_update?: boolean;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailRelation": {
            count?: number;
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailRelationItem"][];
            key?: string;
            label?: string;
            page?: number;
            resource?: string;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailRelationItem": {
            id?: string;
            label?: string;
            subtitle?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ErrorResponse": {
            code?: string;
            conflict?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.WriteConflictResponse"];
            details?: unknown;
            /** @description Error and Fields are kept as compatibility aliases for existing clients. */
            error?: string;
            field_errors?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.FieldErrorResponse"][];
            fields?: {
                [key: string]: string;
            };
            localized_key?: string;
            message?: string;
            request_id?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FacilityJobListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FacilityJobResponse"][];
            next_cursor?: string;
            previous_cursor?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FacilityJobResponse": {
            attempts?: number;
            /** @enum {string} */
            class?: "mutation" | "export";
            completed_at?: string;
            created_at?: string;
            error?: string;
            failure_count?: number;
            job_id?: string;
            /** @enum {string} */
            kind?: "control_cabinet" | "sps_controller" | "sps_controller_system_type" | "field_device" | "object_data";
            processed?: number;
            progress?: number;
            result?: Record<string, never>;
            retryable?: boolean;
            stage?: string;
            /** @enum {string} */
            status?: "queued" | "running" | "completed" | "failed";
            success_count?: number;
            total?: number;
            /** @enum {string} */
            type?: "copy" | "export" | "bulk" | "delete" | "restore";
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceCreateResultResponse": {
            /** @description Error message if failed (empty if succeeded) */
            error?: string;
            /** @description Specific field that caused the error (if applicable) */
            error_field?: string;
            /** @description The created field device (null if failed) */
            field_device?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceResponse"];
            /** @description Index in the original request array */
            index?: number;
            /** @description Whether the creation succeeded */
            success?: boolean;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceDetailResponse": {
            capabilities?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailCapabilities"];
            field_device?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceResponse"];
            relations?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailRelation"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceOptionsResponse": {
            /** @description apparat_id -> [system_part_ids] */
            apparat_to_system_part?: {
                [key: string]: string[];
            };
            apparats?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatResponse"][];
            /** @description object_data_id -> [apparat_ids] */
            object_data_to_apparat?: {
                [key: string]: string[];
            };
            object_datas?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ObjectDataResponse"][];
            system_parts?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemPartResponse"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceResponse": {
            apparat?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatResponse"];
            apparat_id?: string;
            apparat_nr?: number;
            bacnet_objects?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectResponse"][];
            bmk?: string;
            created_at?: string;
            description?: string;
            id?: string;
            specification?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SpecificationResponse"];
            specification_id?: string;
            /** @description Embedded related entities for display */
            sps_controller_system_type?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeResponse"];
            sps_controller_system_type_id?: string;
            system_part?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemPartResponse"];
            system_part_id?: string;
            text_fix?: string;
            updated_at?: string;
            version?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.MultiCreateFieldDeviceRequest": {
            field_devices: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateFieldDeviceRequest"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.MultiCreateFieldDeviceResponse": {
            failure_count?: number;
            results?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceCreateResultResponse"][];
            success_count?: number;
            total_requests?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.NextAvailableGADeviceResponse": {
            ga_device?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.NotificationClassListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.NotificationClassResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.NotificationClassResponse": {
            ack_required_error?: boolean;
            ack_required_normal?: boolean;
            ack_required_not_normal?: boolean;
            created_at?: string;
            event_category?: string;
            id?: string;
            internal_description?: string;
            meaning?: string;
            nc?: number;
            norm_error?: number;
            norm_normal?: number;
            norm_not_normal?: number;
            object_description?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ObjectDataResponse": {
            apparats?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatResponse"][];
            bacnet_objects?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectResponse"][];
            created_at?: string;
            description?: string;
            id?: string;
            is_active?: boolean;
            project_id?: string;
            updated_at?: string;
            version?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalFloat64": {
            set?: boolean;
            /** Format: float64 */
            value?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalInt": {
            set?: boolean;
            value?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString": {
            set?: boolean;
            value?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.PutAlarmValuesRequest": {
            values: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.AlarmValueInput"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SpecificationInput": {
            additional_info_motor_valve?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            additional_info_size?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalInt"];
            additional_information_installation_location?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            electrical_connection_acdc?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            electrical_connection_amperage?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalFloat64"];
            electrical_connection_ph?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalInt"];
            electrical_connection_power?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalFloat64"];
            electrical_connection_rotation?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalInt"];
            specification_brand?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            specification_supplier?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            specification_type?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SpecificationResponse": {
            additional_info_motor_valve?: string;
            additional_info_size?: number;
            additional_information_installation_location?: string;
            created_at?: string;
            electrical_connection_acdc?: string;
            electrical_connection_amperage?: number;
            electrical_connection_ph?: number;
            electrical_connection_power?: number;
            electrical_connection_rotation?: number;
            field_device_id?: string;
            id?: string;
            specification_brand?: string;
            specification_supplier?: string;
            specification_type?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerBulkRequest": {
            ids: string[];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerBulkResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerResponse"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerDetailResponse": {
            capabilities?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailCapabilities"];
            relations?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailRelation"][];
            sps_controller?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerResponse"];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerResponse": {
            control_cabinet_id?: string;
            created_at?: string;
            device_description?: string;
            device_location?: string;
            device_name?: string;
            ga_device?: string;
            gateway?: string;
            id?: string;
            ip_address?: string;
            subnet?: string;
            updated_at?: string;
            version?: number;
            vlan?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeDetailResponse": {
            capabilities?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailCapabilities"];
            relations?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailRelation"][];
            sps_controller_system_type?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeResponse"];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeInput": {
            document_name?: string;
            id?: string;
            number?: number;
            system_type_id: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeResponse": {
            /**
             * @description AggregateVersion is retained for clients generated before the unified
             *     version token was introduced. New clients should use Version.
             */
            aggregate_version?: number;
            created_at?: string;
            document_name?: string;
            field_devices_count?: number;
            id?: string;
            number?: number;
            sps_controller_id?: string;
            /** @description Pre-filled names for display in combobox */
            sps_controller_name?: string;
            system_type_id?: string;
            system_type_name?: string;
            updated_at?: string;
            version?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.StateTextListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.StateTextResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.StateTextResponse": {
            created_at?: string;
            id?: string;
            ref_number?: number;
            state_text1?: string;
            /**
             * @description Include only first few for lightness or all? User just wants search.
             *     But detailed response usually contains all.
             */
            state_text2?: string;
            state_text3?: string;
            state_text4?: string;
            state_text5?: string;
            state_text6?: string;
            state_text7?: string;
            state_text8?: string;
            state_text9?: string;
            state_text10?: string;
            state_text11?: string;
            state_text12?: string;
            state_text13?: string;
            state_text14?: string;
            state_text15?: string;
            state_text16?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemPartListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemPartResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemPartResponse": {
            created_at?: string;
            description?: string;
            id?: string;
            name?: string;
            short_name?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemTypeListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemTypeResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SystemTypeResponse": {
            created_at?: string;
            id?: string;
            name?: string;
            number_max?: number;
            number_min?: number;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UnitListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UnitResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UnitResponse": {
            code?: string;
            id?: string;
            name?: string;
            symbol?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateAlarmDefinitionRequest": {
            alarm_note?: string;
            alarm_type_id?: string;
            name?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateAlarmFieldRequest": {
            data_type?: string;
            default_unit_code?: string;
            key?: string;
            label?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateAlarmTypeFieldRequest": {
            default_unit_id?: string;
            default_value_json?: string;
            display_order?: number;
            is_required?: boolean;
            is_user_editable?: boolean;
            ui_group?: string;
            validation_json?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateAlarmTypeRequest": {
            name?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateApparatRequest": {
            description?: string;
            name?: string;
            short_name?: string;
            system_part_ids?: string[];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateBacnetObjectRequest": {
            alarm_definition_id?: string;
            alarm_type_id?: string;
            base_version?: number;
            description?: string;
            field_device_id?: string;
            gms_visible?: boolean;
            hardware_quantity?: number;
            hardware_type?: string;
            notification_class_id?: string;
            object_data_id?: string;
            optional?: boolean;
            software_number?: number;
            software_reference_id?: string;
            software_type?: string;
            state_text_id?: string;
            text_fix?: string;
            text_individual?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateBuildingRequest": {
            base_version?: number;
            building_group?: number;
            iws_code?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateControlCabinetRequest": {
            base_version?: number;
            building_id?: string;
            control_cabinet_nr?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateFieldDeviceRequest": {
            apparat_id?: string;
            apparat_nr?: number;
            bacnet_objects?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectInput"][];
            base_version?: number;
            bmk?: string;
            description?: string;
            object_data_id?: string;
            sps_controller_system_type_id?: string;
            system_part_id: string;
            text_fix?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateFieldDeviceSpecificationRequest": {
            additional_info_motor_valve?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            additional_info_size?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalInt"];
            additional_information_installation_location?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            electrical_connection_acdc?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            electrical_connection_amperage?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalFloat64"];
            electrical_connection_ph?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalInt"];
            electrical_connection_power?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalFloat64"];
            electrical_connection_rotation?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalInt"];
            specification_brand?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            specification_supplier?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
            specification_type?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.OptionalString"];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateNotificationClassRequest": {
            ack_required_error?: boolean;
            ack_required_normal?: boolean;
            ack_required_not_normal?: boolean;
            event_category?: string;
            internal_description?: string;
            meaning?: string;
            nc?: number;
            norm_error?: number;
            norm_normal?: number;
            norm_not_normal?: number;
            object_description?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateObjectDataRequest": {
            apparat_ids?: string[];
            bacnet_objects?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectInput"][];
            description?: string;
            is_active?: boolean;
            project_id?: string;
            version?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerRequest": {
            base_version?: number;
            control_cabinet_id?: string;
            device_description?: string;
            device_location?: string;
            device_name?: string;
            ga_device?: string;
            gateway?: string;
            ip_address?: string;
            subnet?: string;
            system_types?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeInput"][];
            vlan?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerSystemTypeRequest": {
            base_version?: number;
            document_name?: string;
            number?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateStateTextRequest": {
            ref_number?: number;
            state_text1?: string;
            state_text2?: string;
            state_text3?: string;
            state_text4?: string;
            state_text5?: string;
            state_text6?: string;
            state_text7?: string;
            state_text8?: string;
            state_text9?: string;
            state_text10?: string;
            state_text11?: string;
            state_text12?: string;
            state_text13?: string;
            state_text14?: string;
            state_text15?: string;
            state_text16?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSystemPartRequest": {
            description?: string;
            name?: string;
            short_name?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSystemTypeRequest": {
            name?: string;
            number_max?: number;
            number_min?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateUnitRequest": {
            code?: string;
            name?: string;
            symbol?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ValidateBuildingRequest": {
            building_group?: number;
            id?: string;
            iws_code?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ValidateControlCabinetRequest": {
            building_id?: string;
            control_cabinet_nr?: string;
            id?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ValidateSPSControllerRequest": {
            control_cabinet_id?: string;
            device_name?: string;
            ga_device?: string;
            gateway?: string;
            id?: string;
            ip_address?: string;
            subnet?: string;
            vlan?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ChangeEventResponse": {
            /** @enum {string} */
            action?: "create" | "update" | "delete" | "restore";
            actor_id?: string;
            actor_name?: string;
            after_json?: Record<string, never>;
            batch_id?: string;
            before_json?: Record<string, never>;
            diff_json?: Record<string, never>;
            entity_id?: string;
            entity_table?: string;
            id?: string;
            metadata_json?: Record<string, never>;
            occurred_at?: string;
            scopes?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ScopeResponse"][];
            summary?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ErrorResponse": {
            code?: string;
            conflict?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.WriteConflictResponse"];
            details?: unknown;
            /** @description Error and Fields are kept as compatibility aliases for existing clients. */
            error?: string;
            field_errors?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.FieldErrorResponse"][];
            fields?: {
                [key: string]: string;
            };
            localized_key?: string;
            message?: string;
            request_id?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ScopeResponse": {
            label?: string;
            scope_id?: string;
            scope_type?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_history.TimelineResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ChangeEventResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.ErrorResponse": {
            code?: string;
            conflict?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.WriteConflictResponse"];
            details?: unknown;
            /** @description Error and Fields are kept as compatibility aliases for existing clients. */
            error?: string;
            field_errors?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.FieldErrorResponse"][];
            fields?: {
                [key: string]: string;
            };
            localized_key?: string;
            message?: string;
            request_id?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SendSMTPTestEmailRequest": {
            body?: string;
            subject?: string;
            to: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SMTPSettingsResponse": {
            allow_insecure_tls?: boolean;
            auth_mode?: string;
            enabled?: boolean;
            from_email?: string;
            from_name?: string;
            has_password?: boolean;
            host?: string;
            id?: string;
            port?: number;
            provider?: string;
            reply_to?: string;
            security?: string;
            updated_at?: string;
            updated_by_id?: string;
            username?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SystemNotificationListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SystemNotificationResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
            unread_count?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.SystemNotificationResponse": {
            actor_id?: string;
            body?: string;
            created_at?: string;
            event_key?: string;
            id?: string;
            is_important?: boolean;
            metadata?: {
                [key: string]: string;
            };
            read_at?: string;
            recipient_id?: string;
            resource_id?: string;
            resource_type?: string;
            title?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.UpsertSMTPSettingsRequest": {
            allow_insecure_tls: boolean;
            /** @enum {string} */
            auth_mode: "none" | "plain";
            enabled: boolean;
            from_email: string;
            from_name?: string;
            host: string;
            password?: string;
            port: number;
            reply_to?: string;
            /** @enum {string} */
            security: "none" | "starttls" | "tls";
            username?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.UpsertUserNotificationPreferenceRequest": {
            /** @enum {string} */
            channel: "email" | "system" | "both";
            /** @enum {string} */
            frequency: "immediate" | "hourly" | "daily" | "weekly";
            notification_email?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.UserNotificationPreferenceResponse": {
            channel?: string;
            created_at?: string;
            email_verification_expires_at?: string;
            email_verification_sent_at?: string;
            frequency?: string;
            id?: string;
            notification_email?: string;
            notification_email_verified_at?: string;
            updated_at?: string;
            user_id?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_notification.VerifyUserNotificationEmailRequest": {
            code: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreatePhaseRequest": {
            name: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectControlCabinetRequest": {
            control_cabinet_id: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectFieldDeviceRequest": {
            field_device_id: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectObjectDataRequest": {
            object_data_id: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectRequest": {
            description?: string;
            name: string;
            phase_id: string;
            start_date?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.SwissDateTime"];
            /** @enum {string} */
            status?: "planned" | "ongoing" | "completed";
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectSPSControllerRequest": {
            sps_controller_id: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectUserRequest": {
            user_id: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ErrorResponse": {
            code?: string;
            conflict?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.WriteConflictResponse"];
            details?: unknown;
            /** @description Error and Fields are kept as compatibility aliases for existing clients. */
            error?: string;
            field_errors?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.FieldErrorResponse"][];
            fields?: {
                [key: string]: string;
            };
            localized_key?: string;
            message?: string;
            request_id?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.MultiCreateProjectFieldDeviceRequest": {
            field_device_ids?: string[];
            field_devices?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CreateFieldDeviceRequest"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.MultiCreateProjectFieldDeviceResponse": {
            association_errors?: string[];
            success_field_device_ids?: string[];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ObjectDataListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ObjectDataResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ObjectDataResponse": {
            apparats?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatResponse"][];
            bacnet_objects?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectResponse"][];
            created_at?: string;
            description?: string;
            id?: string;
            is_active?: boolean;
            project_id?: string;
            updated_at?: string;
            version?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.PhaseListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.PhaseResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.PhaseResponse": {
            created_at?: string;
            id?: string;
            name?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectCapabilitiesResponse": {
            permissions?: string[];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectChangeResponse": {
            action?: string;
            actor_id?: string;
            aggregate_id?: string;
            aggregate_type?: string;
            changed_fields?: string[];
            event_id?: string;
            occurred_at?: string;
            parent_refs?: {
                [key: string]: string;
            };
            revision?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectChangesResponse": {
            current_revision?: number;
            events?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectChangeResponse"][];
            has_more?: boolean;
            project_id?: string;
            reset_required?: boolean;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectControlCabinetListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectControlCabinetResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectControlCabinetResponse": {
            control_cabinet_id?: string;
            created_at?: string;
            id?: string;
            project_id?: string;
            updated_at?: string;
            version?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectFieldDeviceListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectFieldDeviceResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectFieldDeviceResponse": {
            created_at?: string;
            field_device_id?: string;
            id?: string;
            project_id?: string;
            updated_at?: string;
            version?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectResponse": {
            created_at?: string;
            creator_id?: string;
            description?: string;
            id?: string;
            name?: string;
            phase?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.PhaseResponse"];
            phase_id?: string;
            start_date?: string;
            status?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_domain_project.ProjectStatus"];
            updated_at?: string;
            version?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectSPSControllerListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectSPSControllerResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectSPSControllerResponse": {
            created_at?: string;
            id?: string;
            project_id?: string;
            sps_controller_id?: string;
            updated_at?: string;
            version?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectUserListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserResponse"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.ProjectUserResponse": {
            project_id?: string;
            user_id?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.SwissDateTime": {
            "time.Time"?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdatePhaseRequest": {
            name?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectControlCabinetRequest": {
            base_version?: number;
            control_cabinet_id: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectFieldDeviceRequest": {
            base_version?: number;
            field_device_id: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectRequest": {
            base_version?: number;
            description?: string;
            name?: string;
            phase_id?: string;
            /** Format: date-time */
            start_date?: string | null;
            /** @enum {unknown} */
            status?: "planned" | "ongoing" | "completed";
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectSPSControllerRequest": {
            base_version?: number;
            sps_controller_id: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_team.AddTeamMemberRequest": {
            /** @enum {string} */
            role: "member" | "manager" | "owner";
            user_id: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_team.CreateTeamRequest": {
            description?: string;
            name: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_team.ErrorResponse": {
            code?: string;
            conflict?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.WriteConflictResponse"];
            details?: unknown;
            /** @description Error and Fields are kept as compatibility aliases for existing clients. */
            error?: string;
            field_errors?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.FieldErrorResponse"][];
            fields?: {
                [key: string]: string;
            };
            localized_key?: string;
            message?: string;
            request_id?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_team.TeamListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.TeamResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_team.TeamMemberListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_team.TeamMemberResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_team.TeamMemberResponse": {
            joined_at?: string;
            role?: string;
            team_id?: string;
            user_id?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_team.TeamResponse": {
            created_at?: string;
            description?: string;
            id?: string;
            name?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_team.UpdateTeamRequest": {
            description?: string;
            name?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.AddRolePermissionRequest": {
            permission: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.AdminSetUserRoleRequest": {
            /** @enum {string} */
            role: "superadmin" | "admin_fzag" | "fzag" | "admin_planer" | "planer" | "admin_entrepreneur" | "entrepreneur";
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.AllowedRole": {
            display_name?: string;
            role?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.AllowedRolesResponse": {
            roles?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.AllowedRole"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.CreatePermissionRequest": {
            action: string;
            description?: string;
            name: string;
            resource: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.CreateUserRequest": {
            created_by_id?: string;
            email: string;
            first_name: string;
            is_active?: boolean;
            last_name: string;
            password: string;
            /** @enum {string} */
            role?: "superadmin" | "admin_fzag" | "fzag" | "admin_planer" | "planer" | "admin_entrepreneur" | "entrepreneur";
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.ErrorResponse": {
            code?: string;
            conflict?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.WriteConflictResponse"];
            details?: unknown;
            /** @description Error and Fields are kept as compatibility aliases for existing clients. */
            error?: string;
            field_errors?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.FieldErrorResponse"][];
            fields?: {
                [key: string]: string;
            };
            localized_key?: string;
            message?: string;
            request_id?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.PermissionResponse": {
            action?: string;
            created_at?: string;
            description?: string;
            id?: string;
            name?: string;
            resource?: string;
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.RegistrationProcessResponse": {
            accepted_at?: string;
            can_resend?: boolean;
            email_status?: string;
            expires_at?: string;
            last_error?: string;
            last_sent_at?: string;
            resend_available_at?: string;
            send_count?: number;
            status?: string;
            steps?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.RegistrationProcessStepResponse"][];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.RegistrationProcessStepResponse": {
            key?: string;
            label?: string;
            status?: string;
            timestamp?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.RolePermissionResponse": {
            created_at?: string;
            id?: string;
            permission?: string;
            role?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_domain_user.Role"];
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.RoleResponse": {
            can_manage?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_domain_user.Role"][];
            created_at?: string;
            description?: string;
            display_name?: string;
            id?: string;
            level?: number;
            name?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_domain_user.Role"];
            permissions?: string[];
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UpdateOwnPasswordRequest": {
            current_password: string;
            new_password: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UpdatePermissionRequest": {
            description?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UpdateRolePermissionsRequest": {
            permissions?: string[];
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UpdateUserRequest": {
            email?: string;
            first_name?: string;
            is_active?: boolean;
            last_name?: string;
            password?: string;
            /** @enum {string} */
            role?: "superadmin" | "admin_fzag" | "fzag" | "admin_planer" | "planer" | "admin_entrepreneur" | "entrepreneur";
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryCapabilitiesResponse": {
            can_change_role?: boolean;
            can_delete?: boolean;
            can_disable?: boolean;
            can_enable?: boolean;
            can_restore?: boolean;
            can_update?: boolean;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryListResponse": {
            capabilities?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryPageCapabilitiesResponse"];
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryUserResponse"][];
            page?: number;
            roles?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryRoleFilterResponse"][];
            teams?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryTeamFilterResponse"][];
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryPageCapabilitiesResponse": {
            can_create_user?: boolean;
            can_read_deleted?: boolean;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryRoleFilterResponse": {
            count?: number;
            display_name?: string;
            role?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryTeamFilterResponse": {
            count?: number;
            id?: string;
            name?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryTeamResponse": {
            id?: string;
            name?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryUserResponse": {
            capabilities?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryCapabilitiesResponse"];
            created_at?: string;
            deleted_at?: string;
            disabled_at?: string;
            email?: string;
            failed_login_attempts?: number;
            first_name?: string;
            id?: string;
            is_active?: boolean;
            is_anonymized?: boolean;
            is_deleted?: boolean;
            last_login_at?: string;
            last_name?: string;
            locked_until?: string;
            registration_process?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.RegistrationProcessResponse"];
            restore_until?: string;
            role?: string;
            role_display_name?: string;
            teams?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserDirectoryTeamResponse"][];
            updated_at?: string;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_user.UserResponse": {
            created_at?: string;
            deleted_at?: string;
            disabled_at?: string;
            email?: string;
            failed_login_attempts?: number;
            first_name?: string;
            id?: string;
            is_active?: boolean;
            is_anonymized?: boolean;
            is_deleted?: boolean;
            last_login_at?: string;
            last_name?: string;
            locked_until?: string;
            restore_until?: string;
            role?: string;
            role_display_name?: string;
            updated_at?: string;
        };
        "internal_handler_facility.ErrorResponse": {
            code?: string;
            conflict?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.WriteConflictResponse"];
            details?: unknown;
            /** @description Error and Fields are kept as compatibility aliases for existing clients. */
            error?: string;
            field_errors?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.FieldErrorResponse"][];
            fields?: {
                [key: string]: string;
            };
            localized_key?: string;
            message?: string;
            request_id?: string;
        };
        "internal_handler_facility.ObjectDataListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ObjectDataResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "internal_handler_facility.ObjectDataResponse": {
            apparats?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ApparatResponse"][];
            bacnet_objects?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BacnetObjectResponse"][];
            created_at?: string;
            description?: string;
            id?: string;
            is_active?: boolean;
            project_id?: string;
            updated_at?: string;
            version?: string;
        };
        "internal_handler_facility.SPSControllerSystemTypeListResponse": {
            items?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeResponse"][];
            page?: number;
            total?: number;
            total_pages?: number;
        };
        "internal_handler_facility.SPSControllerSystemTypeResponse": {
            /**
             * @description AggregateVersion is retained for clients generated before the unified
             *     version token was introduced. New clients should use Version.
             */
            aggregate_version?: number;
            created_at?: string;
            document_name?: string;
            field_devices_count?: number;
            id?: string;
            number?: number;
            sps_controller_id?: string;
            /** @description Pre-filled names for display in combobox */
            sps_controller_name?: string;
            system_type_id?: string;
            system_type_name?: string;
            updated_at?: string;
            version?: number;
        };
        "internal_handler_i18n.ErrorResponse": {
            code?: string;
            conflict?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.WriteConflictResponse"];
            details?: unknown;
            /** @description Error and Fields are kept as compatibility aliases for existing clients. */
            error?: string;
            field_errors?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.FieldErrorResponse"][];
            fields?: {
                [key: string]: string;
            };
            localized_key?: string;
            message?: string;
            request_id?: string;
        };
        "internal_handler_project.ErrorResponse": {
            code?: string;
            conflict?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.WriteConflictResponse"];
            details?: unknown;
            /** @description Error and Fields are kept as compatibility aliases for existing clients. */
            error?: string;
            field_errors?: components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_common.FieldErrorResponse"][];
            fields?: {
                [key: string]: string;
            };
            localized_key?: string;
            message?: string;
            request_id?: string;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: {
        /** @description Building data */
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateBuildingRequest": {
            content: {
                "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateBuildingRequest"];
            };
        };
        /** @description Control Cabinet data */
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateControlCabinetRequest": {
            content: {
                "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateControlCabinetRequest"];
            };
        };
        /** @description Field Device data */
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateFieldDeviceRequest": {
            content: {
                "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateFieldDeviceRequest"];
            };
        };
        /** @description SPS Controller data */
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerRequest": {
            content: {
                "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerRequest"];
            };
        };
        /** @description SPS controller system type data */
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerSystemTypeRequest": {
            content: {
                "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerSystemTypeRequest"];
            };
        };
        /** @description Partial phase data */
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdatePhaseRequest": {
            content: {
                "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdatePhaseRequest"];
            };
        };
        /** @description Partial project data */
        "github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectRequest": {
            content: {
                "application/json": components["schemas"]["github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectRequest"];
            };
        };
    };
    headers: never;
    pathItems: never;
};
export type $defs = Record<string, never>;
export type operations = Record<string, never>;
